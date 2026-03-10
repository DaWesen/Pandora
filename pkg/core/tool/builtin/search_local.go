package builtin

import (
	"encoding/json"
	"fmt"

	"github.com/DaWesen/Pandora/pkg/core"
)

// SearchLocalTool 实现本地搜索工具（不依赖外部 API）
type SearchLocalTool struct{}

// NewSearchLocalTool 创建新的本地搜索工具实例
func NewSearchLocalTool() *SearchLocalTool {
	return &SearchLocalTool{}
}

// Name 返回工具名称
func (t *SearchLocalTool) Name() string {
	return "search"
}

// Description 返回工具描述
func (t *SearchLocalTool) Description() string {
	return "提供本地搜索功能，模拟搜索结果。"
}

// Schema 返回工具参数模式
func (t *SearchLocalTool) Schema() core.ToolSchema {
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

// Execute 执行搜索操作
func (t *SearchLocalTool) Execute(input core.ToolInput) (core.ToolOutput, error) {
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

	// 模拟搜索结果
	results := t.mockSearch(params.Query, params.Num)

	return core.ToolOutput{
		Content: fmt.Sprintf("找到 %d 个关于 '%s' 的搜索结果", len(results), params.Query),
		Data: map[string]any{
			"query":   params.Query,
			"num":     params.Num,
			"results": results,
		},
	}, nil
}

// mockSearch 模拟搜索结果
func (t *SearchLocalTool) mockSearch(query string, num int) []map[string]any {
	// 模拟搜索结果，根据查询关键词生成相关结果
	baseResults := []map[string]any{
		{
			"title":   fmt.Sprintf("%s - 相关信息", query),
			"link":    fmt.Sprintf("https://example.com/search?q=%s", query),
			"snippet": fmt.Sprintf("这是关于 '%s' 的搜索结果，包含相关信息和资源。", query),
		},
		{
			"title":   "Pandora AI Assistant - 智能助手系统",
			"link":    "https://example.com/pandora",
			"snippet": "Pandora 是一个功能强大的智能助手系统，支持工具调用、联网搜索等功能。",
		},
		{
			"title":   "如何使用 Pandora AI Assistant",
			"link":    "https://example.com/pandora/usage",
			"snippet": "本教程将指导您如何使用 Pandora AI Assistant 的各种功能，包括工具调用和联网搜索。",
		},
		{
			"title":   "Pandora AI Assistant 的工具系统",
			"link":    "https://example.com/pandora/tools",
			"snippet": "Pandora AI Assistant 支持多种工具，包括时间查询、文件操作、联网搜索等。",
		},
		{
			"title":   "Pandora AI Assistant 的架构设计",
			"link":    "https://example.com/pandora/architecture",
			"snippet": "本文介绍了 Pandora AI Assistant 的架构设计，包括核心组件和工作原理。",
		},
	}

	// 限制返回结果数量
	mockResults := baseResults
	if len(mockResults) > num {
		mockResults = mockResults[:num]
	}

	return mockResults
}
