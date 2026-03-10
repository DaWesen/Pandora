package builtin

import (
	"encoding/json"
	"testing"

	"github.com/DaWesen/Pandora/pkg/core"
)

func TestWebTool_Name(t *testing.T) {
	tool := NewWebTool()
	if tool.Name() != "web" {
		t.Errorf("Expected name 'web', got '%s'", tool.Name())
	}
}

func TestWebTool_Description(t *testing.T) {
	tool := NewWebTool()
	desc := tool.Description()
	if desc == "" {
		t.Errorf("Expected non-empty description")
	}
}

func TestWebTool_Schema(t *testing.T) {
	tool := NewWebTool()
	schema := tool.Schema()
	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}
	if len(schema.Required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(schema.Required))
	}
}

func TestWebTool_Execute(t *testing.T) {
	// 准备工具输入
	params := map[string]interface{}{
		"url":    "https://httpbin.org/get",
		"method": "GET",
	}
	args, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}
	input := core.ToolInput{Arguments: args}

	// 执行工具
	tool := NewWebTool()
	output, err := tool.Execute(input)
	if err != nil {
		t.Fatalf("Failed to execute web operation: %v", err)
	}

	// 验证结果
	if output.Content == "" {
		t.Errorf("Expected non-empty output content")
	}
}
