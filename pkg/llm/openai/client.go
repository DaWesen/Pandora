package openai

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

type Config struct {
	APIKey      string         `json:"api_key"`
	BaseURL     string         `json:"base_url"`
	Model       string         `json:"model"`
	Timeout     time.Duration  `json:"timeout"`
	HTTPClient  *http.Client   `json:"http_client"`
	ExtraFields map[string]any `json:"extra_fields,omitempty"`
}

type Client struct {
	config     *Config
	httpClient *http.Client
}

// OpenAI的消息结构
type OpenAIMessage struct {
	Role         string              `json:"role"`
	Content      string              `json:"content"`
	ToolCalls    []OpenAIToolCall    `json:"tool_calls,omitempty"`
	ToolCallID   string              `json:"tool_call_id,omitempty"`
	MultiContent []OpenAIMessagePart `json:"multi_content,omitempty"`
}

// OpenAI的消息部分结构（用于多模态）
type OpenAIMessagePart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL struct {
		URL    string `json:"url"`
		Detail string `json:"detail,omitempty"`
	} `json:"image_url,omitempty"`
	AudioURL struct {
		URL      string `json:"url"`
		MIMEType string `json:"mime_type,omitempty"`
	} `json:"audio_url,omitempty"`
	VideoURL struct {
		URL string `json:"url"`
	} `json:"video_url,omitempty"`
}

// OpenAI的工具调用结构
type OpenAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	} `json:"function"`
}

// OpenAI的工具结构
type OpenAITool struct {
	Type     string                 `json:"type"`
	Function map[string]interface{} `json:"function"`
}

// 表示OpenAI的聊天请求结构
type ChatRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
	Tools    []OpenAITool    `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
	// 通用参数
	Temperature      *float32 `json:"temperature,omitempty"`
	TopP             *float32 `json:"top_p,omitempty"`
	MaxTokens        *int     `json:"max_tokens,omitempty"`
	Stop             []string `json:"stop,omitempty"`
	PresencePenalty  *float32 `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float32 `json:"frequency_penalty,omitempty"`
	// 额外字段
	ExtraFields map[string]any `json:"extra_fields,omitempty"`
}

// 表示OpenAI的聊天相应结构
type ChatResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []Choice       `json:"choices"`
	Usage   map[string]int `json:"usage,omitempty"`
}

// 表示OpenAI响应中的选择
type Choice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// 表示OpenAI流式响应中的增量
type Delta struct {
	Role             string           `json:"role,omitempty"`
	Content          string           `json:"content,omitempty"`
	ToolCalls        []OpenAIToolCall `json:"tool_calls,omitempty"`
	ReasoningContent string           `json:"reasoning_content,omitempty"`
}

// 表示 OpenAI API 的流式响应结构
type StreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []Delta        `json:"choices"`
	Usage   map[string]int `json:"usage,omitempty"`
}

// 创建新的OpenAI客户端
func NewClient(config *Config) *Client {
	if config == nil {
		panic("config cannot be nil")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	if config.Timeout == 0 {
		config.Timeout = 120 * time.Second
	}

	var httpClient *http.Client
	if config.HTTPClient != nil {
		httpClient = config.HTTPClient
	} else {
		httpClient = &http.Client{Timeout: config.Timeout}
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
	}
}

// 实现 core.LLMClient 接口的 Chat 方法
func (c *Client) Chat(messages []core.Message, tools []core.Tool) (core.Message, error) {
	openAIMsgs := make([]OpenAIMessage, len(messages))
	for i, msg := range messages {
		openAIMsgs[i] = c.toOpenAIMessage(msg)
	}

	// 构建请求
	reqBody := ChatRequest{
		Model:    c.config.Model,
		Messages: openAIMsgs,
		Stream:   false,
	}

	// 添加工具
	if len(tools) > 0 {
		reqBody.Tools = c.toOpenAITools(tools)
	}

	// 添加额外字段
	if c.config.ExtraFields != nil {
		reqBody.ExtraFields = c.config.ExtraFields
	}

	// 序列化请求体
	requestJSON, err := json.Marshal(reqBody)
	if err != nil {
		return core.Message{}, err
	}

	// 创建请求
	req, err := http.NewRequest("POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(requestJSON))
	if err != nil {
		return core.Message{}, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return core.Message{}, err
	}
	defer resp.Body.Close()

	// 读取响应
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return core.Message{}, err
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return core.Message{}, fmt.Errorf("API error: %s", string(bodyBytes))
	}

	// 解析响应
	var chatResp ChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return core.Message{}, err
	}

	// 检查是否有选择
	if len(chatResp.Choices) == 0 {
		return core.Message{}, fmt.Errorf("no choices in response")
	}

	// 转换为 core.Message
	return c.toCoreMessage(chatResp.Choices[0].Message), nil
}

// 实现 core.LLMClient 接口的 ChatStream 方法
func (c *Client) ChatStream(messages []core.Message, tools []core.Tool) (<-chan core.Message, <-chan error) {
	msgChan := make(chan core.Message)
	errChan := make(chan error)

	// 在 goroutine 中处理流式请求
	go func() {
		defer close(msgChan)
		defer close(errChan)

		openAIMsgs := make([]OpenAIMessage, len(messages))
		for i, msg := range messages {
			openAIMsgs[i] = c.toOpenAIMessage(msg)
		}

		// 构建请求
		reqBody := ChatRequest{
			Model:    c.config.Model,
			Messages: openAIMsgs,
			Stream:   true,
		}

		// 添加工具
		if len(tools) > 0 {
			reqBody.Tools = c.toOpenAITools(tools)
		}

		// 添加额外字段
		if c.config.ExtraFields != nil {
			reqBody.ExtraFields = c.config.ExtraFields
		}

		// 序列化请求体
		requestJSON, err := json.Marshal(reqBody)
		if err != nil {
			errChan <- err
			return
		}

		// 创建请求
		req, err := http.NewRequest("POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(requestJSON))
		if err != nil {
			errChan <- err
			return
		}

		// 设置请求头
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

		// 发送请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("API error: %s", string(bodyBytes))
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

			// 跳过空行和事件流前缀
			line = strings.TrimSpace(line)
			if line == "" || line == "data: [DONE]" {
				continue
			}

			// 移除 "data: " 前缀
			line = strings.TrimPrefix(line, "data: ")

			// 解析 JSON 响应
			var streamResp StreamResponse
			if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
				errChan <- err
				return
			}

			// 检查是否有选择
			if len(streamResp.Choices) == 0 {
				continue
			}

			// 构建消息
			delta := streamResp.Choices[0]
			if delta.Content != "" || len(delta.ToolCalls) > 0 || delta.ReasoningContent != "" {
				msg := core.Message{
					Role:    core.Role(delta.Role),
					Content: delta.Content,
				}

				// 转换工具调用
				if len(delta.ToolCalls) > 0 {
					msg.ToolCalls = make([]core.ToolCall, len(delta.ToolCalls))
					for i, tc := range delta.ToolCalls {
						msg.ToolCalls[i] = core.ToolCall{
							ID:   tc.ID,
							Type: tc.Type,
						}
						msg.ToolCalls[i].Function.Name = tc.Function.Name
						msg.ToolCalls[i].Function.Arguments = tc.Function.Arguments
					}
				}

				// 处理推理内容
				if delta.ReasoningContent != "" {
					if msg.Metadata == nil {
						msg.Metadata = make(map[string]any)
					}
					msg.Metadata["reasoning_content"] = delta.ReasoningContent
				}

				//发送消息
				msgChan <- msg
			}
		}
	}()

	return msgChan, errChan
}

// 转换 core.Message 为 OpenAIMessage
func (c *Client) toOpenAIMessage(msg core.Message) OpenAIMessage {
	openAIMsg := OpenAIMessage{
		Role:       string(msg.Role),
		Content:    msg.Content,
		ToolCallID: msg.ToolCallID,
	}

	// 转换工具调用
	if len(msg.ToolCalls) > 0 {
		openAIMsg.ToolCalls = make([]OpenAIToolCall, len(msg.ToolCalls))
		for i, tc := range msg.ToolCalls {
			openAIMsg.ToolCalls[i] = OpenAIToolCall{
				ID:   tc.ID,
				Type: tc.Type,
			}
			openAIMsg.ToolCalls[i].Function.Name = tc.Function.Name
			openAIMsg.ToolCalls[i].Function.Arguments = tc.Function.Arguments
		}
	}

	// 处理多模态内容
	if len(msg.Images) > 0 || msg.Audio != nil || msg.Video != nil {
		openAIMsg.MultiContent = make([]OpenAIMessagePart, 0)

		// 处理图片
		for _, img := range msg.Images {
			part := OpenAIMessagePart{
				Type: "image_url",
			}
			part.ImageURL.URL = img.Data
			if img.Format != "" {
				part.ImageURL.Detail = img.Format
			}
			openAIMsg.MultiContent = append(openAIMsg.MultiContent, part)
		}

		// 处理音频
		if msg.Audio != nil {
			part := OpenAIMessagePart{
				Type: "audio_url",
			}
			part.AudioURL.URL = msg.Audio.Data
			part.AudioURL.MIMEType = msg.Audio.Format
			openAIMsg.MultiContent = append(openAIMsg.MultiContent, part)
		}

		// 处理视频
		if msg.Video != nil {
			part := OpenAIMessagePart{
				Type: "video_url",
			}
			part.VideoURL.URL = msg.Video.Data
			openAIMsg.MultiContent = append(openAIMsg.MultiContent, part)
		}
	}

	return openAIMsg
}

// 转换 OpenAIMessage 为 core.Message
func (c *Client) toCoreMessage(openAIMsg OpenAIMessage) core.Message {
	msg := core.Message{
		Role:       core.Role(openAIMsg.Role),
		Content:    openAIMsg.Content,
		ToolCallID: openAIMsg.ToolCallID,
	}

	// 转换工具调用
	if len(openAIMsg.ToolCalls) > 0 {
		msg.ToolCalls = make([]core.ToolCall, len(openAIMsg.ToolCalls))
		for i, tc := range openAIMsg.ToolCalls {
			msg.ToolCalls[i] = core.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
			}
			msg.ToolCalls[i].Function.Name = tc.Function.Name
			msg.ToolCalls[i].Function.Arguments = tc.Function.Arguments
		}
	}

	// 处理多模态内容
	if len(openAIMsg.MultiContent) > 0 {
		for _, part := range openAIMsg.MultiContent {
			switch part.Type {
			case "image_url":
				img := core.ImageContent{
					Type:   "image",
					Data:   part.ImageURL.URL,
					Format: part.ImageURL.Detail,
				}
				msg.Images = append(msg.Images, img)
			case "audio_url":
				if msg.Audio == nil {
					msg.Audio = &core.AudioContent{
						Type:   "audio",
						Data:   part.AudioURL.URL,
						Format: part.AudioURL.MIMEType,
					}
				}
			case "video_url":
				if msg.Video == nil {
					msg.Video = &core.VideoContent{
						Type: "video",
						Data: part.VideoURL.URL,
					}
				}
			}
		}
	}

	return msg
}

// 转换 core.Tool 为 OpenAITool
func (c *Client) toOpenAITools(tools []core.Tool) []OpenAITool {
	openAITools := make([]OpenAITool, len(tools))
	for i, tool := range tools {
		schema := tool.Schema()
		openAITools[i] = OpenAITool{
			Type: "function",
			Function: map[string]interface{}{
				"name":        tool.Name(),
				"description": tool.Description(),
				"parameters": map[string]interface{}{
					"type":       schema.Type,
					"properties": schema.Properties,
					"required":   schema.Required,
				},
			},
		}
	}
	return openAITools
}
