package builtin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DaWesen/Pandora/pkg/core"
)

type WebTool struct{}

// 创建联网工具实例
func NewWebTool() *WebTool {
	return &WebTool{}
}

func (t *WebTool) Name() string {
	return "web"
}

func (t *WebTool) Description() string {
	return "提供联网功能，发送 HTTP 请求并获取响应。"
}

// 返回工具参数模式
func (t *WebTool) Schema() core.ToolSchema {
	return core.ToolSchema{
		Type: "object",
		Properties: map[string]core.Property{
			"url": {
				Type:        "string",
				Description: "请求的 URL",
			},
			"method": {
				Type:        "string",
				Description: "HTTP 方法，如 GET、POST、PUT、DELETE 等",
				Enum:        []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
			},
			"headers": {
				Type:        "object",
				Description: "HTTP 请求头，如 {\"Content-Type\": \"application/json\"}",
			},
			"body": {
				Type:        "string",
				Description: "HTTP 请求体，仅在 POST、PUT、PATCH 等方法时使用",
			},
			"timeout": {
				Type:        "integer",
				Description: "请求超时时间（秒），默认为 30",
			},
		},
		Required: []string{"url", "method"},
	}
}

// 执行联网操作
func (t *WebTool) Execute(input core.ToolInput) (core.ToolOutput, error) {
	var params struct {
		URL     string            `json:"url"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers,omitempty"`
		Body    string            `json:"body,omitempty"`
		Timeout int               `json:"timeout,omitempty"`
	}
	if err := json.Unmarshal(input.Arguments, &params); err != nil {
		return core.ToolOutput{}, err
	}

	if params.Timeout <= 0 {
		params.Timeout = 30
	}

	//创建http客户端
	client := &http.Client{
		Timeout: time.Duration(params.Timeout) * time.Second,
	}

	//创建请求
	req, err := http.NewRequest(params.Method, params.URL, strings.NewReader(params.Body))
	if err != nil {
		return core.ToolOutput{}, err
	}
	if params.Headers != nil {
		for key, value := range params.Headers {
			req.Header.Set(key, value)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return core.ToolOutput{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return core.ToolOutput{}, err
	}

	// 构建响应
	response := map[string]any{
		"status_code": resp.StatusCode,
		"status":      resp.Status,
		"headers":     resp.Header,
		"body":        string(body),
	}

	//解析
	var jsonBody any
	if err := json.Unmarshal(body, &jsonBody); err == nil {
		response["json"] = jsonBody
	}
	return core.ToolOutput{
		Content: fmt.Sprintf("HTTP %s 请求 %s 完成，状态码: %d", params.Method, params.URL, resp.StatusCode),
		Data:    response,
	}, nil
}
