package system

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

var githubUsernamePattern = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$`)

func ValidGitHubUsername(name string) bool {
	return githubUsernamePattern.MatchString(name) && !strings.Contains(name, "--")
}

type GitHubUser struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	Bio         string `json:"bio"`
	Location    string `json:"location"`
	Company     string `json:"company"`
	AvatarURL   string `json:"avatar_url"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	CreatedAt   string `json:"created_at"`
}
type GitHubRepository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	UpdatedAt   string `json:"updated_at"`
	Private     bool   `json:"private"`
	Fork        bool   `json:"fork"`
	Archived    bool   `json:"archived"`
}
type GitHubCommit struct {
	SHA        string `json:"sha"`
	Message    string `json:"message"`
	Date       string `json:"date"`
	Repository string `json:"repository"`
}
type GitHubEvent struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Repository string `json:"repository"`
	CreatedAt  string `json:"createdAt"`
}
type GitHubProfile struct {
	Username     string             `json:"username"`
	Profile      *GitHubUser        `json:"profile"`
	Repositories []GitHubRepository `json:"repositories"`
	Commits      []GitHubCommit     `json:"commits"`
	Events       []GitHubEvent      `json:"events"`
	Warnings     []string           `json:"warnings"`
	FetchedAt    string             `json:"fetchedAt"`
}
type githubCacheItem struct {
	data  *GitHubProfile
	until time.Time
}
type githubReader struct {
	client  *http.Client
	base    string
	mu      sync.Mutex
	cache   map[string]githubCacheItem
	flights singleflight.Group
}

var publicGitHubReader = &githubReader{
	client: &http.Client{Timeout: 12 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return errors.New("不跟随 GitHub API 重定向") }},
	base:   "https://api.github.com", cache: map[string]githubCacheItem{},
}

func ReadGitHubProfile(ctx context.Context, username string) (*GitHubProfile, error) {
	return publicGitHubReader.read(ctx, username)
}
func (r *githubReader) request(ctx context.Context, path string, output any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.base+path, nil)
	if err != nil {
		return errors.New("GitHub 请求构建失败")
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "GBlog-Public-Profile")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	res, err := r.client.Do(req)
	if err != nil {
		return errors.New("无法连接 GitHub，请检查服务器网络或稍后重试")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusNotFound:
		return errors.New("GitHub 账号或资源不存在")
	case http.StatusUnauthorized:
		return errors.New("GitHub 数据访问未获授权")
	case http.StatusForbidden, http.StatusTooManyRequests:
		return errors.New("GitHub 限流或权限不足，请稍后重试")
	}
	if res.StatusCode != http.StatusOK {
		return errors.New("GitHub 暂时无法提供数据")
	}
	if err = json.NewDecoder(io.LimitReader(res.Body, 2<<20)).Decode(output); err != nil {
		return errors.New("GitHub 返回数据无法解析")
	}
	return nil
}
func (r *githubReader) read(parent context.Context, username string) (*GitHubProfile, error) {
	if !ValidGitHubUsername(username) {
		return nil, errors.New("GitHub 用户名格式无效")
	}
	now := time.Now().UTC()
	key := strings.ToLower(username)
	r.mu.Lock()
	cached, ok := r.cache[key]
	r.mu.Unlock()
	if ok && now.Before(cached.until) {
		return cached.data, nil
	}
	result := r.flights.DoChan(key, func() (any, error) {
		r.mu.Lock()
		cached, ok := r.cache[key]
		r.mu.Unlock()
		if ok && time.Now().Before(cached.until) {
			return cached.data, nil
		}
		return r.fetch(username, key)
	})
	select {
	case <-parent.Done():
		return nil, errors.New("GitHub 请求已取消")
	case value := <-result:
		if value.Err != nil {
			return nil, value.Err
		}
		return value.Val.(*GitHubProfile), nil
	}
}
func (r *githubReader) fetch(username, key string) (*GitHubProfile, error) {
	now := time.Now().UTC()
	// Shared fetch remains bounded even when the initiating tab is closed.
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	profile := &GitHubUser{}
	if err := r.request(ctx, "/users/"+username, profile); err != nil {
		return nil, err
	}
	result := &GitHubProfile{Username: profile.Login, Profile: profile, Repositories: []GitHubRepository{}, Commits: []GitHubCommit{}, Events: []GitHubEvent{}, Warnings: []string{}, FetchedAt: now.Format(time.RFC3339)}
	var repos []GitHubRepository
	if err := r.request(ctx, "/users/"+username+"/repos?type=owner&sort=updated&direction=desc&per_page=12", &repos); err != nil {
		result.Warnings = append(result.Warnings, "公开仓库："+err.Error())
	} else {
		for _, repo := range repos {
			if !repo.Private {
				result.Repositories = append(result.Repositories, repo)
			}
		}
	}
	var commits struct {
		Items []struct {
			SHA        string `json:"sha"`
			Repository struct {
				FullName string `json:"full_name"`
				Private  bool   `json:"private"`
			} `json:"repository"`
			Commit struct {
				Message string `json:"message"`
				Author  struct {
					Date string `json:"date"`
				} `json:"author"`
			} `json:"commit"`
		} `json:"items"`
	}
	if err := r.request(ctx, "/search/commits?q="+url.QueryEscape("author:"+username+" is:public")+"&sort=committer-date&order=desc&per_page=5", &commits); err != nil {
		result.Warnings = append(result.Warnings, "近期公开提交："+err.Error())
	} else {
		for _, item := range commits.Items {
			if !item.Repository.Private {
				result.Commits = append(result.Commits, GitHubCommit{SHA: item.SHA, Repository: item.Repository.FullName, Message: strings.SplitN(item.Commit.Message, "\n", 2)[0], Date: item.Commit.Author.Date})
			}
		}
	}
	var events []struct {
		ID        string `json:"id"`
		Type      string `json:"type"`
		Public    bool   `json:"public"`
		CreatedAt string `json:"created_at"`
		Repo      struct {
			Name string `json:"name"`
		} `json:"repo"`
	}
	if err := r.request(ctx, "/users/"+username+"/events/public?per_page=10", &events); err != nil {
		result.Warnings = append(result.Warnings, "近期公开动态："+err.Error())
	} else {
		for _, event := range events {
			if event.Public {
				result.Events = append(result.Events, GitHubEvent{ID: event.ID, Type: event.Type, Repository: event.Repo.Name, CreatedAt: event.CreatedAt})
			}
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.cache) >= 128 {
		r.cache = map[string]githubCacheItem{}
	}
	ttl := 10 * time.Minute
	if len(result.Warnings) > 0 {
		ttl = 2 * time.Minute
	}
	r.cache[key] = githubCacheItem{data: result, until: time.Now().Add(ttl)}
	return result, nil
}
