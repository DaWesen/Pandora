package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DaWesen/Pandora/pkg/core"
)

type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
	timeout    time.Duration
}

type OllamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []OllamaToolCall `json:"tool_calls,omitempty"`
}

// OllamaToolCall 表示 Ollama API 的工具调用结构
type OllamaToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	} `json:"function"`
}

type OllamaTool struct {
	Type     string      `json:"type"` // 固定为 "function"
	Function FunctionDef `json:"function"`
}

type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  core.ToolSchema `json:"parameters"`
}

// 与 core.ToolCall 中的结构保持一致
type Function struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ChatRequest 表示 Ollama API 的聊天请求结构
type ChatRequest struct {
	Model    string                 `json:"model"`
	Messages []OllamaMessage        `json:"messages"`
	Tools    []OllamaTool           `json:"tools,omitempty"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// ChatResponse 表示 Ollama API 的聊天响应结构
type ChatResponse struct {
	Model              string        `json:"model"`
	CreatedAt          string        `json:"created_at"`
	Message            OllamaMessage `json:"message"`
	Done               bool          `json:"done"`
	TotalDuration      float64       `json:"total_duration"`
	LoadDuration       float64       `json:"load_duration"`
	PromptEvalCount    int           `json:"prompt_eval_count"`
	PromptEvalDuration float64       `json:"prompt_eval_duration"`
	EvalCount          int           `json:"eval_count"`
	EvalDuration       float64       `json:"eval_duration"`
}

// 创建ollama客户端
func NewClient(baseURL, model string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &Client{
		baseURL:    baseURL,
		model:      model,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		timeout:    120 * time.Second,
	}
}

// 实现core.LLMClient接口
func (c *Client) Chat(messages []core.Message, tools []core.Tool) (core.Message, error) {
	ollamaMsgs := make([]OllamaMessage, len(messages))
	for i, msg := range messages {
		ollamaMsgs[i] = c.toOllamaMessage(msg)
	}
	var ollamaTools []OllamaTool
	if len(tools) > 0 {
		ollamaTools = c.toOllamaTools(tools)
	}
	reqBody := ChatRequest{
		Model:    c.model,
		Messages: ollamaMsgs,
		Tools:    ollamaTools,
		Stream:   false,
		Options: map[string]interface{}{
			"temperature":    0.7,
			"num_predict":    2048,
			"top_p":          0.9,
			"repeat_penalty": 1.1,
		},
	}

	// 序列化请求体
	requestJSON, err := json.Marshal(reqBody)
	if err != nil {
		return core.Message{}, err
	}

	// 发送 HTTP 请求
	resp, err := c.httpClient.Post(
		c.baseURL+"/api/chat",
		"application/json",
		bytes.NewBuffer(requestJSON),
	)
	if err != nil {
		return core.Message{}, err
	}
	defer resp.Body.Close()

	// 解析响应
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return core.Message{}, err
	}

	// 转换为 core.Message
	return c.toCoreMessage(chatResp.Message), nil
}

// 实现流式对话接口
func (c *Client) ChatStream(messages []core.Message, tools []core.Tool) (<-chan core.Message, <-chan error) {
	msgChan := make(chan core.Message)
	errChan := make(chan error)

	// 在 goroutine 中处理流式请求
	go func() {
		defer close(msgChan)
		defer close(errChan)

		// 构建请求
		ollamaMsgs := make([]OllamaMessage, len(messages))
		for i, msg := range messages {
			ollamaMsgs[i] = c.toOllamaMessage(msg)
		}
		var ollamaTools []OllamaTool
		if len(tools) > 0 {
			ollamaTools = c.toOllamaTools(tools)
		}
		reqBody := ChatRequest{
			Model:    c.model,
			Messages: ollamaMsgs,
			Tools:    ollamaTools,
			Stream:   true, // 启用流式
			Options: map[string]interface{}{
				"temperature":    0.7,
				"num_predict":    2048,
				"top_p":          0.9,
				"repeat_penalty": 1.1,
			},
		}

		// 序列化请求体
		requestJSON, err := json.Marshal(reqBody)
		if err != nil {
			errChan <- err
			return
		}

		// 创建请求
		req, err := http.NewRequest("POST",
			c.baseURL+"/api/chat",
			bytes.NewBuffer(requestJSON))
		if err != nil {
			errChan <- err
			return
		}
		req.Header.Set("Content-Type", "application/json")

		// 发送请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode != http.StatusOK {
			errChan <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			return
		}

		// 读取并处理流式响应
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				errChan <- err
				return
			}

			// 跳过空行
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// 解析 JSON 响应
			var chatResp ChatResponse
			if err := json.Unmarshal([]byte(line), &chatResp); err != nil {
				errChan <- err
				return
			}

			// 发送消息到通道
			if chatResp.Message.Content != "" || len(chatResp.Message.ToolCalls) > 0 {
				msgChan <- c.toCoreMessage(chatResp.Message)
			}

			// 检查是否完成
			if chatResp.Done {
				break
			}
		}
	}()

	return msgChan, errChan
}

// toCoreMessage 转换 OllamaMessage 为 core.Message
func (c *Client) toCoreMessage(om OllamaMessage) core.Message {
	msg := core.Message{
		Role:    core.Role(om.Role),
		Content: om.Content,
	}

	// 转换工具调用
	if len(om.ToolCalls) > 0 {
		msg.ToolCalls = make([]core.ToolCall, len(om.ToolCalls))
		for i, tc := range om.ToolCalls {
			msg.ToolCalls[i] = core.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
			}
			msg.ToolCalls[i].Function.Name = tc.Function.Name
			msg.ToolCalls[i].Function.Arguments = tc.Function.Arguments
		}
	}

	return msg
}

// 转化core.Message为ollama消息格式
func (c *Client) toOllamaMessage(msg core.Message) OllamaMessage {
	om := OllamaMessage{
		Role:    string(msg.Role),
		Content: msg.Content,
	}
	if len(msg.ToolCalls) > 0 {
		om.ToolCalls = make([]OllamaToolCall, len(msg.ToolCalls))
		for i, tc := range msg.ToolCalls {
			om.ToolCalls[i] = OllamaToolCall{
				ID:   tc.ID,
				Type: tc.Type,
			}
			om.ToolCalls[i].Function.Name = tc.Function.Name
			om.ToolCalls[i].Function.Arguments = tc.Function.Arguments
		}
	}
	return om
}

// toOllamaTools 转换core.Tool → ollama.Tool格式
func (c *Client) toOllamaTools(tools []core.Tool) []OllamaTool {
	result := make([]OllamaTool, len(tools))
	for i, tool := range tools {
		schema := tool.Schema()
		result[i] = OllamaTool{
			Type: "function",
			Function: FunctionDef{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  schema,
			},
		}
	}
	return result
}
