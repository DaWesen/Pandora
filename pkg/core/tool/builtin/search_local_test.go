package builtin

import (
	"encoding/json"
	"testing"

	"github.com/DaWesen/Pandora/pkg/core"
)

func TestSearchLocalTool_Name(t *testing.T) {
	tool := NewSearchLocalTool()
	if tool.Name() != "search" {
		t.Errorf("Expected name 'search', got '%s'", tool.Name())
	}
}

func TestSearchLocalTool_Description(t *testing.T) {
	tool := NewSearchLocalTool()
	desc := tool.Description()
	if desc == "" {
		t.Errorf("Expected non-empty description")
	}
}

func TestSearchLocalTool_Schema(t *testing.T) {
	tool := NewSearchLocalTool()
	schema := tool.Schema()
	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}
	if len(schema.Required) != 1 {
		t.Errorf("Expected 1 required field, got %d", len(schema.Required))
	}
}

func TestSearchLocalTool_Execute(t *testing.T) {
	// 准备工具输入
	params := map[string]interface{}{
		"query": "Pandora AI Assistant",
		"num":   3,
	}
	args, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}
	input := core.ToolInput{Arguments: args}

	// 执行工具
	tool := NewSearchLocalTool()
	output, err := tool.Execute(input)
	if err != nil {
		t.Fatalf("Failed to execute search operation: %v", err)
	}

	// 验证结果
	if output.Content == "" {
		t.Errorf("Expected non-empty output content")
	}

	// 验证结果数量
	data, ok := output.Data["results"].([]map[string]any)
	if !ok {
		t.Errorf("Expected results to be a slice of maps")
	}
	if len(data) != 3 {
		t.Errorf("Expected 3 results, got %d", len(data))
	}
}
