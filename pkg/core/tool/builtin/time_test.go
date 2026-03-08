package builtin

import (
	"encoding/json"
	"testing"

	"github.com/DaWesen/Pandora/pkg/core"
)

func TestTimeTool(t *testing.T) {
	tool := NewTimeTool()

	// 测试基本信息
	if tool.Name() != "time" {
		t.Errorf("Expected tool name 'time', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Tool description should not be empty")
	}

	// 测试 Schema
	schema := tool.Schema()
	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}

	// 测试 Execute 方法
	testCases := []struct {
		name     string
		args     map[string]any
		expected bool // 预期是否成功
	}{
		{
			name:     "Full format",
			args:     map[string]any{"format": "full"},
			expected: true,
		},
		{
			name:     "Date format",
			args:     map[string]any{"format": "date"},
			expected: true,
		},
		{
			name:     "Time format",
			args:     map[string]any{"format": "time"},
			expected: true,
		},
		{
			name:     "Timestamp format",
			args:     map[string]any{"format": "timestamp"},
			expected: true,
		},
		{
			name:     "With timezone",
			args:     map[string]any{"format": "full", "timezone": "UTC"},
			expected: true,
		},
		{
			name:     "Invalid timezone",
			args:     map[string]any{"format": "full", "timezone": "Invalid/Timezone"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			argsJSON, err := json.Marshal(tc.args)
			if err != nil {
				t.Fatalf("Failed to marshal args: %v", err)
			}

			input := core.ToolInput{
				Arguments: argsJSON,
			}

			output, err := tool.Execute(input)

			if tc.expected {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if output.Content == "" {
					t.Error("Expected non-empty content")
				}
			} else {
				if err == nil {
					t.Error("Expected error, got none")
				}
			}
		})
	}
}
