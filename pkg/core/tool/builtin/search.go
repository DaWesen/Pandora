package builtin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/DaWesen/Pandora/pkg/core"
)

// 实现通用搜索引擎工具
type SearchTool struct{}

// NewSearchTool 创建新的搜索工具实例
func NewSearchTool() *SearchTool {
	return &SearchTool{}
}

// 返回工具名称
func (t *SearchTool) Name() string {
	return "search"
}

// 返回工具描述
func (t *SearchTool) Description() string {
	return "提供通用的联网搜索功能，获取互联网上的信息。"
}

// 返回工具参数模式
func (t *SearchTool) Schema() core.ToolSchema {
	return core.ToolSchema{
		Type: "object",
		Properties: map[string]core.Property{
			"query": {
				Type:        "string",
				Description: "搜索查询关键词",
			},
			"num": {
				Type:        "integer",
				Description: "返回结果数量，默认为 5",
			},
		},
		Required: []string{"query"},
	}
}

// 执行搜索操作
func (t *SearchTool) Execute(input core.ToolInput) (core.ToolOutput, error) {
	var params struct {
		Query string `json:"query"`
		Num   int    `json:"num,omitempty"`
	}

	if err := json.Unmarshal(input.Arguments, &params); err != nil {
		return core.ToolOutput{}, err
	}

	// 设置默认值
	if params.Num <= 0 {
		params.Num = 5
	}

	// 执行搜索
	results, err := t.search(params.Query, params.Num)
	if err != nil {
		return core.ToolOutput{}, err
	}

	return core.ToolOutput{
		Content: fmt.Sprintf("找到 %d 个关于 '%s' 的搜索结果", len(results), params.Query),
		Data: map[string]any{
			"query":   params.Query,
			"num":     params.Num,
			"results": results,
		},
	}, nil
}

// 执行实际的搜索操作
func (t *SearchTool) search(query string, num int) ([]map[string]any, error) {
	// 使用 DuckDuckGo API
	baseURL := "https://api.duckduckgo.com"
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("q", query)
	q.Set("format", "json")
	q.Set("no_html", "1")
	q.Set("no_redirect", "1")
	u.RawQuery = q.Encode()

	// 发送请求
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		Abstract      string `json:"Abstract"`
		AbstractURL   string `json:"AbstractURL"`
		Heading       string `json:"Heading"`
		RelatedTopics []struct {
			FirstURL string `json:"FirstURL"`
			Text     string `json:"Text"`
			Result   string `json:"Result"`
		} `json:"RelatedTopics"`
		Results []struct {
			FirstURL string `json:"FirstURL"`
			Text     string `json:"Text"`
			Result   string `json:"Result"`
		} `json:"Results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 格式化结果
	var results []map[string]any

	// 添加摘要结果
	if result.Abstract != "" {
		results = append(results, map[string]any{
			"title":   result.Heading,
			"link":    result.AbstractURL,
			"snippet": result.Abstract,
		})
	}

	// 添加相关主题结果
	for _, topic := range result.RelatedTopics {
		if topic.FirstURL != "" && topic.Text != "" {
			// 提取标题和摘要
			title := topic.Text
			snippet := topic.Text
			if idx := strings.Index(title, " - "); idx != -1 {
				title = title[:idx]
				snippet = title[idx+3:]
			}

			results = append(results, map[string]any{
				"title":   title,
				"link":    topic.FirstURL,
				"snippet": snippet,
			})
		}
	}

	// 添加常规结果
	for _, item := range result.Results {
		if item.FirstURL != "" && item.Text != "" {
			// 提取标题和摘要
			title := item.Text
			snippet := item.Text
			if idx := strings.Index(title, " - "); idx != -1 {
				title = title[:idx]
				snippet = title[idx+3:]
			}

			results = append(results, map[string]any{
				"title":   title,
				"link":    item.FirstURL,
				"snippet": snippet,
			})
		}
	}

	// 限制返回结果数量
	if len(results) > num {
		results = results[:num]
	}

	return results, nil
}
