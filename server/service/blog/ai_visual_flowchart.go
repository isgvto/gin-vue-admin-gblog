package blog

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var compactNode = regexp.MustCompile(`\b([A-Za-z0-9_]+)\s*(?:\(\[|\(\(|\{\{|\[\[|\[\(|\[|\(|\{)\s*"([^"]*)"\s*(?:\]\)|\)\)|\}\}|\]\]|\)\]|\]|\)|\})`)
var compactTextEdge = regexp.MustCompile(`--\s*"[^"]*"\s*-->`)
var compactInlineClass = regexp.MustCompile(`:::[A-Za-z0-9_]+`)
var compactQuoted = regexp.MustCompile(`"[^"]*"`)
var compactEdgeLabel = regexp.MustCompile(`\|[^|\r\n]*\|`)
var compactWord = regexp.MustCompile(`[A-Za-z0-9_]+`)
var compactArrow = regexp.MustCompile(`-->|---|==>|-\.->`)

// 仅约束AI自动生成的流程图，不限制作者已有的Mermaid文档或结构图。
func validateCompactFlowchart(source string) error {
	nodes := map[string]bool{}
	groups, depth := 0, 0
	for _, match := range compactNode.FindAllStringSubmatch(source, -1) {
		if utf8.RuneCountInString(strings.Trim(match[2], "`")) > 36 {
			return errors.New("流程图节点文字超过36字，请缩短标签后重新生成")
		}
	}
	source = compactNode.ReplaceAllString(source, "$1")
	source = compactTextEdge.ReplaceAllString(source, "-->")
	source = compactQuoted.ReplaceAllString(source, "")
	source = compactInlineClass.ReplaceAllString(source, "")
	source = compactEdgeLabel.ReplaceAllString(source, "")
	var statements []string
	for _, line := range strings.Split(strings.ReplaceAll(source, ";", "\n"), "\n") {
		line = strings.TrimSpace(strings.SplitN(line, "%%", 2)[0])
		if fields := strings.Fields(line); len(fields) > 0 && fields[0] == "subgraph" {
			groups++
			depth++
			if depth > 1 || groups > 4 {
				return errors.New("流程图最多4个阶段分组，不嵌套分组；请调整布局或按阶段拆分")
			}
			continue
		}
		if line == "end" {
			depth--
			if depth < 0 {
				return errors.New("流程图分组未正确闭合")
			}
			continue
		}
		skip := false
		for _, prefix := range []string{"flowchart ", "graph ", "direction ", "classDef ", "class ", "style ", "linkStyle "} {
			if strings.HasPrefix(line, prefix) {
				skip = true
				break
			}
		}
		if skip || line == "" {
			continue
		}
		statements = append(statements, line)
		for _, id := range compactWord.FindAllString(line, -1) {
			nodes[id] = true
		}
	}
	body := strings.Join(statements, "\n")
	if depth != 0 {
		return errors.New("流程图分组未正确闭合")
	}
	if len(nodes) == 0 || len(nodes) > 24 || len(compactArrow.FindAllString(body, -1)) > 40 {
		return errors.New("流程图超过安全上限（24个节点、40条连线），请按阶段拆分后生成")
	}
	return nil
}
