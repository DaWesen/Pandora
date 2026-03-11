package builtin

import (
	"encoding/json"
	"testing"

	"github.com/DaWesen/Pandora/pkg/core"
)

func TestCrawlerTool_Name(t *testing.T) {
	tool := NewCrawlerTool()
	if tool.Name() != "crawler" {
		t.Errorf("Expected name 'crawler', got '%s'", tool.Name())
	}
}

func TestCrawlerTool_Description(t *testing.T) {
	tool := NewCrawlerTool()
	desc := tool.Description()
	if desc == "" {
		t.Errorf("Expected non-empty description")
	}
}

func TestCrawlerTool_Schema(t *testing.T) {
	tool := NewCrawlerTool()
	schema := tool.Schema()
	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}
	if len(schema.Required) != 1 {
		t.Errorf("Expected 1 required field, got %d", len(schema.Required))
	}
}

func TestCrawlerTool_Execute(t *testing.T) {
	// 准备工具输入
	params := map[string]interface{}{
		"url":     "https://httpbin.org",
		"extract": "all",
	}
	args, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}
	input := core.ToolInput{Arguments: args}

	// 执行工具
	tool := NewCrawlerTool()
	output, err := tool.Execute(input)
	if err != nil {
		t.Fatalf("Failed to execute crawler operation: %v", err)
	}

	// 验证结果
	if output.Content == "" {
		t.Errorf("Expected non-empty output content")
	}

	// 验证数据
	if output.Data == nil {
		t.Errorf("Expected non-nil data")
	}
}
