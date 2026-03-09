package builtin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/DaWesen/Pandora/pkg/core"
)

func TestFileTool_Name(t *testing.T) {
	tool := NewFileTool()
	if tool.Name() != "filetool" {
		t.Errorf("Expected name 'filetool', got '%s'", tool.Name())
	}
}

func TestFileTool_Description(t *testing.T) {
	tool := NewFileTool()
	desc := tool.Description()
	if desc == "" {
		t.Errorf("Expected non-empty description")
	}
}

func TestFileTool_Schema(t *testing.T) {
	tool := NewFileTool()
	schema := tool.Schema()
	if schema.Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", schema.Type)
	}
	if len(schema.Required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(schema.Required))
	}
}

func TestFileTool_ReadFile(t *testing.T) {
	// 创建测试文件
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.txt")
	testContent := "Hello, World!"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 准备工具输入
	params := map[string]interface{}{
		"operation": "read",
		"path":      testFile,
	}
	args, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}
	input := core.ToolInput{Arguments: args}

	// 执行工具
	tool := NewFileTool()
	output, err := tool.Execute(input)
	if err != nil {
		t.Fatalf("Failed to execute read operation: %v", err)
	}

	// 验证结果
	if output.Content != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, output.Content)
	}
}

func TestFileTool_WriteFile(t *testing.T) {
	// 创建测试目录
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test.txt")
	testContent := "Hello, World!"

	// 准备工具输入
	params := map[string]interface{}{
		"operation": "write",
		"path":      testFile,
		"content":   testContent,
		"overwrite": true,
	}
	args, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}
	input := core.ToolInput{Arguments: args}

	// 执行工具
	tool := NewFileTool()
	output, err := tool.Execute(input)
	if err != nil {
		t.Fatalf("Failed to execute write operation: %v", err)
	}

	// 验证结果
	if output.Content == "" {
		t.Errorf("Expected non-empty output content")
	}

	// 验证文件内容
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("Expected file content '%s', got '%s'", testContent, string(content))
	}
}

func TestFileTool_ListDirectory(t *testing.T) {
	// 创建测试目录和文件
	testDir := t.TempDir()
	testFile1 := filepath.Join(testDir, "file1.txt")
	testFile2 := filepath.Join(testDir, "file2.txt")
	if err := os.WriteFile(testFile1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create test file1: %v", err)
	}
	if err := os.WriteFile(testFile2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create test file2: %v", err)
	}

	// 准备工具输入
	params := map[string]interface{}{
		"operation": "list",
		"path":      testDir,
	}
	args, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}
	input := core.ToolInput{Arguments: args}

	// 执行工具
	tool := NewFileTool()
	output, err := tool.Execute(input)
	if err != nil {
		t.Fatalf("Failed to execute list operation: %v", err)
	}

	// 验证结果
	if output.Content == "" {
		t.Errorf("Expected non-empty output content")
	}
}
