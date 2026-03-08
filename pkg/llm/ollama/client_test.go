package ollama

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DaWesen/Pandora/pkg/core"
)

// 测试 NewClient 函数
func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:11434", "llama3")
	if client.baseURL != "http://localhost:11434" {
		t.Errorf("Expected baseURL to be 'http://localhost:11434', got '%s'", client.baseURL)
	}
	if client.model != "llama3" {
		t.Errorf("Expected model to be 'llama3', got '%s'", client.model)
	}
	if client.timeout != 120*time.Second {
		t.Errorf("Expected timeout to be 120s, got '%v'", client.timeout)
	}
}

// 测试消息转换
func TestMessageConversion(t *testing.T) {
	client := NewClient("", "")

	// 测试 toOllamaMessage
	coreMsg := core.Message{
		Role:    core.RoleUser,
		Content: "Hello, world!",
	}
	ollamaMsg := client.toOllamaMessage(coreMsg)
	if ollamaMsg.Role != "user" {
		t.Errorf("Expected role to be 'user', got '%s'", ollamaMsg.Role)
	}
	if ollamaMsg.Content != "Hello, world!" {
		t.Errorf("Expected content to be 'Hello, world!', got '%s'", ollamaMsg.Content)
	}

	// 测试 toCoreMessage
	ollamaMsg2 := OllamaMessage{
		Role:    "assistant",
		Content: "Hi there!",
	}
	coreMsg2 := client.toCoreMessage(ollamaMsg2)
	if coreMsg2.Role != core.RoleAssistant {
		t.Errorf("Expected role to be 'assistant', got '%v'", coreMsg2.Role)
	}
	if coreMsg2.Content != "Hi there!" {
		t.Errorf("Expected content to be 'Hi there!', got '%s'", coreMsg2.Content)
	}
}

// 测试工具转换
func TestToolConversion(t *testing.T) {
	client := NewClient("", "")

	// 创建测试工具
	testTool := &mockTool{
		name:        "test_tool",
		description: "A test tool",
		schema: core.ToolSchema{
			Type: "object",
			Properties: map[string]core.Property{
				"param1": {
					Type:        "string",
					Description: "First parameter",
				},
			},
			Required: []string{"param1"},
		},
	}

	tools := []core.Tool{testTool}
	ollamaTools := client.toOllamaTools(tools)

	if len(ollamaTools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(ollamaTools))
	}

	if ollamaTools[0].Type != "function" {
		t.Errorf("Expected tool type to be 'function', got '%s'", ollamaTools[0].Type)
	}

	if ollamaTools[0].Function.Name != "test_tool" {
		t.Errorf("Expected tool name to be 'test_tool', got '%s'", ollamaTools[0].Function.Name)
	}
}

// 测试非流式对话
func TestChat(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求
		if r.URL.Path != "/api/chat" {
			t.Errorf("Expected path to be '/api/chat', got '%s'", r.URL.Path)
		}

		// 返回模拟响应
		response := ChatResponse{
			Model: "llama3",
			Message: OllamaMessage{
				Role:    "assistant",
				Content: "Hello from Ollama!",
			},
			Done: true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(server.URL, "llama3")

	// 准备测试消息
	messages := []core.Message{
		{
			Role:    core.RoleUser,
			Content: "Hello",
		},
	}

	// 测试 Chat 方法
	response, err := client.Chat(messages, nil)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if response.Role != core.RoleAssistant {
		t.Errorf("Expected role to be 'assistant', got '%v'", response.Role)
	}

	if response.Content != "Hello from Ollama!" {
		t.Errorf("Expected content to be 'Hello from Ollama!', got '%s'", response.Content)
	}
}

// 模拟工具实现
type mockTool struct {
	name        string
	description string
	schema      core.ToolSchema
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Description() string {
	return m.description
}

func (m *mockTool) Schema() core.ToolSchema {
	return m.schema
}

func (m *mockTool) Execute(input core.ToolInput) (core.ToolOutput, error) {
	return core.ToolOutput{
		Content: "Tool executed successfully",
	}, nil
}
