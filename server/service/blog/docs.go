package blog

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/isgvto/gin-vue-admin-gblog/server/global"
	blogModel "github.com/isgvto/gin-vue-admin-gblog/server/model/blog"
	"github.com/isgvto/gin-vue-admin-gblog/server/utils"
	"gorm.io/gorm"
)

type DocsService struct {
	cacheMu sync.RWMutex
	cache   *docsTreeCache
	syncMu  sync.Mutex
}

type docsTreeCache struct {
	key       string
	tree      []DocNode
	fetchedAt time.Time
}

type DocNode struct {
	Title    string    `json:"title"`
	Path     string    `json:"path"`
	Type     string    `json:"type"`
	Children []DocNode `json:"children,omitempty"`
}

type DocContent struct {
	Title   string `json:"title"`
	Path    string `json:"path"`
	Content string `json:"content"`
	RawURL  string `json:"rawUrl"`
	EditURL string `json:"editUrl"`
}

type docsConfig struct {
	Repo   string
	Owner  string
	Name   string
	Branch string
	Root   string
}

type githubTreeResponse struct {
	Tree []githubTreeItem `json:"tree"`
}

type githubTreeItem struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type githubContentResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

type docsManifest struct {
	Key      string                     `json:"key"`
	Tree     []DocNode                  `json:"tree"`
	Docs     map[string]docsManifestDoc `json:"docs"`
	SyncedAt time.Time                  `json:"syncedAt"`
}

type docsManifestDoc struct {
	Title   string `json:"title"`
	Path    string `json:"path"`
	RawURL  string `json:"rawUrl"`
	EditURL string `json:"editUrl"`
}

const (
	docsRepoSetting   = "docsGithubRepo"
	docsBranchSetting = "docsGithubBranch"
	docsRootSetting   = "docsGithubRoot"
	defaultDocsBranch = "main"
	docsSecretSetting = "docsGithubWebhookSecret"
)

var (
	errDocsWebhookSecretNotConfigured = errors.New("docs webhook secret is not configured")
	errDocsWebhookSignatureInvalid    = errors.New("docs webhook signature is invalid")
)

func (s *DocsService) GetTree() ([]DocNode, error) {
	cfg, err := s.getConfig()
	if err != nil {
		return nil, err
	}
	key := cfg.cacheKey()

	s.cacheMu.RLock()
	if s.cache != nil && s.cache.key == key {
		tree := cloneDocNodes(s.cache.tree)
		s.cacheMu.RUnlock()
		return tree, nil
	}
	s.cacheMu.RUnlock()

	manifest, err := s.loadManifest(cfg)
	if err != nil {
		return nil, err
	}
	s.cacheMu.Lock()
	s.cache = &docsTreeCache{
		key:       cfg.cacheKey(),
		tree:      cloneDocNodes(manifest.Tree),
		fetchedAt: manifest.SyncedAt,
	}
	s.cacheMu.Unlock()
	return cloneDocNodes(manifest.Tree), nil
}

func (s *DocsService) SyncTree() ([]DocNode, error) {
	s.syncMu.Lock()
	defer s.syncMu.Unlock()

	cfg, err := s.getConfig()
	if err != nil {
		return nil, err
	}
	tree, docPaths, err := s.fetchTree(cfg)
	if err != nil {
		return nil, err
	}
	manifest := docsManifest{
		Key:      cfg.cacheKey(),
		Tree:     cloneDocNodes(tree),
		Docs:     make(map[string]docsManifestDoc, len(docPaths)),
		SyncedAt: time.Now(),
	}
	cacheRoot := cfg.cacheRoot()
	tempRoot := cacheRoot + fmt.Sprintf(".tmp-%d", time.Now().UnixNano())
	if err := os.RemoveAll(tempRoot); err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempRoot)

	for _, docPath := range docPaths {
		markdown, err := s.fetchMarkdown(cfg, docPath)
		if err != nil {
			return nil, fmt.Errorf("sync %s failed: %w", docPath, err)
		}
		html := utils.MarkdownToHTML(markdown)
		html = rewriteRelativeAssets(html, cfg.rawDirURL(docPath))
		if err := writeCacheFile(tempRoot, "raw", docPath, []byte(markdown)); err != nil {
			return nil, err
		}
		if err := writeCacheFile(tempRoot, "html", strings.TrimSuffix(docPath, path.Ext(docPath))+".html", []byte(html)); err != nil {
			return nil, err
		}
		manifest.Docs[docPath] = docsManifestDoc{
			Title:   extractMarkdownTitle(markdown, docPath),
			Path:    docPath,
			RawURL:  cfg.rawURL(docPath),
			EditURL: cfg.editURL(docPath),
		}
	}
	if err := writeManifest(tempRoot, manifest); err != nil {
		return nil, err
	}
	if err := os.RemoveAll(cacheRoot); err != nil {
		return nil, err
	}
	if err := os.Rename(tempRoot, cacheRoot); err != nil {
		return nil, err
	}

	s.cacheMu.Lock()
	s.cache = &docsTreeCache{
		key:       cfg.cacheKey(),
		tree:      cloneDocNodes(tree),
		fetchedAt: time.Now(),
	}
	s.cacheMu.Unlock()

	return tree, nil
}

func (s *DocsService) GetContent(docPath string) (DocContent, error) {
	cfg, err := s.getConfig()
	if err != nil {
		return DocContent{}, err
	}
	docPath = cleanDocPath(docPath)
	if docPath == "" || strings.Contains(docPath, "..") || !strings.HasSuffix(strings.ToLower(docPath), ".md") {
		return DocContent{}, errors.New("invalid document path")
	}

	manifest, err := s.loadManifest(cfg)
	if err != nil {
		return DocContent{}, err
	}
	doc, ok := manifest.Docs[docPath]
	if !ok {
		return DocContent{}, errors.New("document is not synced")
	}
	htmlPath := strings.TrimSuffix(docPath, path.Ext(docPath)) + ".html"
	body, err := readCacheFile(cfg.cacheRoot(), "html", htmlPath)
	if err != nil {
		return DocContent{}, err
	}

	return DocContent{
		Title:   doc.Title,
		Path:    doc.Path,
		Content: string(body),
		RawURL:  doc.RawURL,
		EditURL: doc.EditURL,
	}, nil
}

func (s *DocsService) ValidateWebhookSignature(body []byte, signature string) error {
	secret := strings.TrimSpace(getSiteSettingValue(docsSecretSetting))
	if secret == "" {
		return errDocsWebhookSecretNotConfigured
	}
	signature = strings.TrimSpace(signature)
	if !strings.HasPrefix(signature, "sha256=") {
		return errDocsWebhookSignatureInvalid
	}
	rawSignature := strings.TrimPrefix(signature, "sha256=")
	expectedMAC := hmac.New(sha256.New, []byte(secret))
	expectedMAC.Write(body)
	expected := hex.EncodeToString(expectedMAC.Sum(nil))
	if !hmac.Equal([]byte(rawSignature), []byte(expected)) {
		return errDocsWebhookSignatureInvalid
	}
	return nil
}

func (s *DocsService) getConfig() (docsConfig, error) {
	if err := ensureDefaultSiteSettings(); err != nil {
		return docsConfig{}, err
	}
	repo := strings.TrimSpace(getSiteSettingValue(docsRepoSetting))
	if repo == "" {
		return docsConfig{}, errors.New("docs github repo is not configured")
	}
	owner, name, err := parseGithubRepo(repo)
	if err != nil {
		return docsConfig{}, err
	}
	branch := strings.TrimSpace(getSiteSettingValue(docsBranchSetting))
	if branch == "" {
		branch = defaultDocsBranch
	}
	root := cleanDocPath(getSiteSettingValue(docsRootSetting))

	return docsConfig{
		Repo:   repo,
		Owner:  owner,
		Name:   name,
		Branch: branch,
		Root:   root,
	}, nil
}

func (s *DocsService) loadManifest(cfg docsConfig) (docsManifest, error) {
	body, err := os.ReadFile(filepath.Join(cfg.cacheRoot(), "manifest.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return docsManifest{}, errors.New("documents are not synced yet")
		}
		return docsManifest{}, err
	}
	var manifest docsManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return docsManifest{}, err
	}
	if manifest.Key != cfg.cacheKey() {
		return docsManifest{}, errors.New("documents cache does not match current config")
	}
	if manifest.Docs == nil {
		manifest.Docs = map[string]docsManifestDoc{}
	}
	return manifest, nil
}

func (s *DocsService) fetchTree(cfg docsConfig) ([]DocNode, []string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", cfg.Owner, cfg.Name, url.PathEscape(cfg.Branch))
	body, err := githubGet(apiURL)
	if err != nil {
		return nil, nil, err
	}

	var resp githubTreeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, nil, err
	}

	root := DocNode{Title: "root", Type: "dir"}
	docPaths := make([]string, 0)
	for _, item := range resp.Tree {
		if item.Type != "blob" || !strings.HasSuffix(strings.ToLower(item.Path), ".md") {
			continue
		}
		if cfg.Root != "" {
			if item.Path != cfg.Root && !strings.HasPrefix(item.Path, cfg.Root+"/") {
				continue
			}
		}
		relativePath := strings.TrimPrefix(item.Path, cfg.Root)
		relativePath = strings.TrimPrefix(relativePath, "/")
		if relativePath == "" {
			continue
		}
		insertDocNode(&root, item.Path, relativePath)
		docPaths = append(docPaths, item.Path)
	}
	sortDocNodes(root.Children)
	sort.Strings(docPaths)
	return root.Children, docPaths, nil
}

func (s *DocsService) fetchMarkdown(cfg docsConfig, docPath string) (string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", cfg.Owner, cfg.Name, escapePath(docPath), url.QueryEscape(cfg.Branch))
	body, err := githubGet(apiURL)
	if err != nil {
		return "", err
	}
	var resp githubContentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	if resp.Encoding != "base64" {
		return "", fmt.Errorf("unsupported github content encoding: %s", resp.Encoding)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(resp.Content, "\n", ""))
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func insertDocNode(root *DocNode, fullPath, relativePath string) {
	parts := strings.Split(relativePath, "/")
	current := root
	for i, part := range parts {
		isFile := i == len(parts)-1
		nodeType := "dir"
		title := part
		nodePath := ""
		if isFile {
			nodeType = "file"
			title = strings.TrimSuffix(part, path.Ext(part))
			nodePath = fullPath
		}

		idx := -1
		for j := range current.Children {
			if current.Children[j].Title == title && current.Children[j].Type == nodeType {
				idx = j
				break
			}
		}
		if idx < 0 {
			current.Children = append(current.Children, DocNode{
				Title: title,
				Path:  nodePath,
				Type:  nodeType,
			})
			idx = len(current.Children) - 1
		}
		current = &current.Children[idx]
	}
}

func sortDocNodes(nodes []DocNode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].Type != nodes[j].Type {
			return nodes[i].Type == "dir"
		}
		return strings.ToLower(nodes[i].Title) < strings.ToLower(nodes[j].Title)
	})
	for i := range nodes {
		sortDocNodes(nodes[i].Children)
	}
}

func getSiteSettingValue(name string) string {
	var row blogModel.SiteSetting
	err := global.GVA_DB.Where("name_en = ?", name).First(&row).Error
	if err != nil || row.Value == nil {
		if err != nil && err != gorm.ErrRecordNotFound {
			global.GVA_LOG.Warn("get site setting failed")
		}
		return ""
	}
	return *row.Value
}

func parseGithubRepo(value string) (string, string, error) {
	value = strings.TrimSpace(strings.TrimSuffix(value, ".git"))
	value = strings.TrimSuffix(value, "/")
	if strings.HasPrefix(value, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(value, "git@github.com:"), "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return parts[0], parts[1], nil
		}
	}
	if !strings.Contains(value, "://") {
		parts := strings.Split(value, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return parts[0], parts[1], nil
		}
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host != "github.com" {
		return "", "", errors.New("invalid github repo")
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("invalid github repo")
	}
	return parts[0], parts[1], nil
}

func githubGet(rawURL string) ([]byte, error) {
	client := &http.Client{Timeout: 45 * time.Second}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gin-vue-admin-gblog-docs")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github request failed: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func (c docsConfig) cacheKey() string {
	return c.Owner + "/" + c.Name + "#" + c.Branch + ":" + c.Root
}

func (c docsConfig) cacheRoot() string {
	sum := sha256.Sum256([]byte(c.cacheKey()))
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	return filepath.Join(wd, "runtime", "docs-cache", hex.EncodeToString(sum[:])[:16])
}

func (c docsConfig) rawURL(docPath string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", c.Owner, c.Name, c.Branch, escapePath(docPath))
}

func (c docsConfig) rawDirURL(docPath string) string {
	dir := path.Dir(docPath)
	if dir == "." {
		dir = ""
	}
	if dir == "" {
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/", c.Owner, c.Name, c.Branch)
	}
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s/", c.Owner, c.Name, c.Branch, escapePath(dir))
}

func (c docsConfig) editURL(docPath string) string {
	return fmt.Sprintf("https://github.com/%s/%s/edit/%s/%s", c.Owner, c.Name, c.Branch, escapePath(docPath))
}

func escapePath(value string) string {
	parts := strings.Split(strings.Trim(value, "/"), "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func cleanDocPath(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = strings.Trim(value, "/")
	if value == "" {
		return ""
	}
	if value == "." {
		return ""
	}
	return path.Clean(value)
}

func extractMarkdownTitle(markdown, docPath string) string {
	for _, line := range strings.Split(markdown, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	base := path.Base(docPath)
	return strings.TrimSuffix(base, path.Ext(base))
}

func rewriteRelativeAssets(html, baseURL string) string {
	replacer := strings.NewReplacer(
		`src="./`, `src="`+baseURL,
		`src="../`, `src="`+baseURL+`../`,
		`src="assets/`, `src="`+baseURL+`assets/`,
		`href="./`, `href="`+baseURL,
		`href="../`, `href="`+baseURL+`../`,
	)
	return replacer.Replace(html)
}

func writeManifest(cacheRoot string, manifest docsManifest) error {
	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cacheRoot, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cacheRoot, "manifest.json"), body, 0644)
}

func writeCacheFile(cacheRoot, kind, docPath string, body []byte) error {
	filePath, err := cacheFilePath(cacheRoot, kind, docPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	return os.WriteFile(filePath, body, 0644)
}

func readCacheFile(cacheRoot, kind, docPath string) ([]byte, error) {
	filePath, err := cacheFilePath(cacheRoot, kind, docPath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filePath)
}

func cacheFilePath(cacheRoot, kind, docPath string) (string, error) {
	docPath = cleanDocPath(docPath)
	if docPath == "" || strings.Contains(docPath, "..") {
		return "", errors.New("invalid cache path")
	}
	base := filepath.Join(cacheRoot, kind)
	filePath := filepath.Join(base, filepath.FromSlash(docPath))
	rel, err := filepath.Rel(base, filePath)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", errors.New("invalid cache path")
	}
	return filePath, nil
}

func cloneDocNodes(nodes []DocNode) []DocNode {
	result := make([]DocNode, len(nodes))
	for i := range nodes {
		result[i] = nodes[i]
		result[i].Children = cloneDocNodes(nodes[i].Children)
	}
	return result
}
