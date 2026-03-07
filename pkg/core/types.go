package core

import "encoding/json"

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// 消息结构
type Message struct {
	//基础字段
	Role    Role   `json:"role"`
	Content string `json:"content"`
	//工具相关
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`

	//多模态拓展
	Images []ImageContent `json:"images,omitempty"`
	Audio  *AudioContent  `json:"audio,omitempty"`
	Video  *VideoContent  `json:"video,omitempty"`

	//元数据
	Metadata  map[string]any `json:"metadata,omitempty"`
	Timestamp int64          `json:"timestamp,omitempty"`
}

// LLM请求调用工具
type ToolCall struct {
	ID   string `json:"id"`
	Type string `json:"type"`

	Function struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	} `json:"function"`
}

type ImageContent struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	Format   string `json:"format"`
	Mimetype string `json:"mimetype"`
}

type AudioContent struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	Format   string `json:"format"`
	Duration int    `json:"duration"` // 音频时长，单位为秒
}

type VideoContent struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	Format   string `json:"format"`
	Duration int    `json:"duration"` // 视频时长，单位为秒
}

// Tool工具接口
type Tool interface {
	Name() string
	Description() string

	// Schema 返回工具的参数模式定义，用于告诉LLM工具的输入参数结构
	Schema() ToolSchema

	// Execute 执行工具逻辑，接收工具输入，返回工具输出和可能的错误
	Execute(input ToolInput) (ToolOutput, error)
}

// 工具输入
type ToolInput struct {
	Arguments json.RawMessage `json:"arguments"`
	Metadata  map[string]any  `json:"metadata,omitempty"`
}

// 工具输出
type ToolOutput struct {
	Content string         `json:"content"`
	Data    map[string]any `json:"data,omitempty"`
}

// Tool工具参数模式定义
type ToolSchema struct {
	Type string `json:"type"`
	//以下字段根据Type不同而不同
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// OpenAIFunction 表示OpenAI函数调用格式
type OpenAIFunction struct {
	Name        string                 `json:"name"`        // 函数名称
	Description string                 `json:"description"` // 函数描述
	Parameters  map[string]interface{} `json:"parameters"`  // 函数参数结构
}

// OllamaTool 表示Ollama工具调用格式
type OllamaTool struct {
	Type     string                 `json:"type"`     // 工具类型，固定为 "function"
	Function map[string]interface{} `json:"function"` // 函数信息
}

// 单个参数属性
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	//其他字段根据Type不同而不同
	Enum []string `json:"enum,omitempty"`
}

// 记忆接口
type Memory interface {
	//添加记忆
	Add(msg Message) error

	//获取最近n条记忆
	GetRecent(n int) ([]Message, error)

	//根据查询获取相关记忆
	Query(query string, n int) ([]Message, error)

	//获取所有记忆
	GetAll() ([]Message, error)

	//清除记忆
	Clear() error
}

// LLM客户端接口
type LLMClient interface {
	//非流式对话
	Chat(messages []Message, tools []Tool) (Message, error)

	//流式对话，返回一个消息通道
	ChatStream(messages []Message, tools []Tool) (<-chan Message, <-chan error)
}

// Agent接口
type Agent interface {
	//Run一次完整的对话流程
	Run(input string) (*Message, error)

	//上下文管理
	AddMessage(msg Message) error
	ClearContext() error

	//工具管理
	RegisterTool(tool Tool) error
}
