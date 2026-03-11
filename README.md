# Pandora

Pandora 是一个功能强大的 AI 代理框架，基于 Go 语言开发，支持多种 LLM 后端、工具执行和记忆管理等核心功能。

## 特性

- **模块化设计**：清晰的分层结构，易于扩展和集成
- **多 LLM 支持**：同时支持 OpenAI 和 Ollama 后端
- **工具系统**：内置文件操作、网络搜索、爬虫等实用工具
- **记忆管理**：支持短期记忆，可扩展长期记忆
- **流式对话**：支持实时响应，提升用户体验
- **多模态支持**：支持图像、音频、视频输入

## 快速开始

### 环境要求

- Go 1.20+
- 支持的 LLM 后端：
  - OpenAI API 密钥（可选）
  - Ollama 服务（可选）

### 安装

```bash
go get github.com/DaWesen/Pandora
```

### 基本使用

#### 1. 配置文件

复制示例配置文件并修改：

```bash
cp config/config.yaml.example config/config.yaml
```

编辑 `config/config.yaml` 文件，配置 LLM 后端信息。

#### 2. 创建基本代理

```go
package main

import (
	"fmt"

	"github.com/DaWesen/Pandora/pkg/core"
	"github.com/DaWesen/Pandora/pkg/core/memory"
	"github.com/DaWesen/Pandora/pkg/core/tool"
	"github.com/DaWesen/Pandora/pkg/core/tool/builtin"
	"github.com/DaWesen/Pandora/pkg/llm/openai"
)

func main() {
	// 创建 LLM 客户端
	llmClient := openai.NewClient("your-openai-api-key")

	// 创建记忆系统
	shortTermMemory := memory.NewShortTermMemory()

	// 创建工具注册表
	toolRegistry := tool.NewRegistry()

	// 注册内置工具
	fileTool := builtin.NewFileTool()
	toolRegistry.Register(fileTool)

	// 创建 Agent
	agent := core.NewAgent(
		llmClient,
		shortTermMemory,
		toolRegistry,
		"You are a helpful assistant.",
		10, // 最大上下文大小
	)

	// 执行对话
	response, err := agent.Run("Hello, what can you do?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Response: %s\n", response.Content)
}
```

#### 3. 使用流式对话

```go
func main() {
	// ... 初始化代码同上 ...

	// 执行流式对话
	msgChan, errChan := agent.StreamRun("Tell me a story about AI")

	// 处理流式响应
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return // 通道关闭
			}
			fmt.Printf("%s", msg.Content)
		case err, ok := <-errChan:
			if !ok {
				return // 通道关闭
			}
			fmt.Printf("Error: %v\n", err)
			return
		}
	}
}
```

## 核心组件

### Chat 包

Chat 包提供了连续对话功能，允许用户与 Agent 进行持续的交互，而不需要每次都显式调用 `Run` 方法。

#### 主要功能

- **连续对话循环**：支持用户与 Agent 的持续交互
- **流式对话**：支持实时响应的流式对话
- **自定义配置**：可自定义提示信息和退出命令

#### 主要方法

- `ChatLoop(agent *core.AgentImpl, prompt, exitCommand string)`: 启动一个连续对话循环
- `StreamingChatLoop(agent *core.AgentImpl, prompt, exitCommand string)`: 启动一个流式连续对话循环

#### 使用示例

```go
// 导入 chat 包
import "github.com/DaWesen/Pandora/pkg/chat"

// 创建 Agent
agent := createAgent()

// 启动连续对话循环
// 使用默认提示和退出命令
chat.ChatLoop(agent, "", "")

// 自定义提示和退出命令
// chat.ChatLoop(agent, "> ", "quit")

// 使用流式对话
// chat.StreamingChatLoop(agent, "", "")
```

### Agent 框架

Agent 是框架的核心组件，负责协调 LLM、记忆和工具系统，实现智能对话和任务执行。

#### 工作原理

Agent 的工作流程如下：

1. **接收输入**：接收用户输入的文本
2. **记忆管理**：将用户输入添加到记忆系统中
3. **上下文构建**：从记忆中获取最近的对话历史，构建对话上下文
4. **LLM 调用**：将上下文和系统提示发送给 LLM，获取响应
5. **工具执行**：如果 LLM 响应包含工具调用，执行相应工具并获取结果
6. **结果处理**：将工具执行结果发送给 LLM，获取最终响应
7. **记忆更新**：将最终响应添加到记忆系统中
8. **返回结果**：将最终响应返回给用户

#### 主要方法

- `Run(input string) (*Message, error)`: 执行一次完整的对话流程
- `StreamRun(input string) (<-chan Message, <-chan error)`: 执行流式对话，实时返回响应
- `AddMessage(msg Message) error`: 直接添加消息到记忆中
- `ClearContext() error`: 清除所有记忆
- `RegisterTool(tool Tool) error`: 注册工具到工具库
- `SystemPrompt() string`: 获取系统提示

#### 创建和配置 Agent

```go
// 创建 Agent 的完整示例
func createAgent() *core.AgentImpl {
	// 1. 选择 LLM 后端
	// 选项 1: 使用 OpenAI
	openaiClient := openai.NewClient("your-openai-api-key")
	
	// 选项 2: 使用 Ollama
	// ollamaClient := ollama.NewClient("http://localhost:11434", "llama3")

	// 2. 创建记忆系统
	memory := memory.NewShortTermMemory()

	// 3. 创建工具注册表
	toolRegistry := tool.NewRegistry()

	// 4. 注册内置工具
	fileTool := builtin.NewFileTool()
	searchTool := builtin.NewSearchTool()
	crawlerTool := builtin.NewCrawlerTool()
	timeTool := builtin.NewTimeTool()
	webTool := builtin.NewWebTool()

	toolRegistry.BatchRegister([]core.Tool{
		fileTool,
		searchTool,
		crawlerTool,
		timeTool,
		webTool,
	})

	// 5. 创建 Agent
	agent := core.NewAgent(
		openaiClient,        // LLM 客户端
		memory,             // 记忆系统
		toolRegistry,       // 工具注册表
		"You are a helpful assistant.",  // 系统提示
		10,                 // 最大上下文大小
	)

	return agent
}
```

#### 高级配置

**系统提示定制**：
```go
// 为特定任务定制系统提示
systemPrompt := `You are a professional software developer. 
You help users with coding problems and provide detailed explanations.`
	agent := core.NewAgent(
		llmClient,
		memory,
		toolRegistry,
		systemPrompt,
		10,
	)
```

**上下文大小调整**：
```go
// 根据 LLM 的上下文窗口大小调整
// GPT-4 支持更大的上下文
agent := core.NewAgent(
	llmClient,
	memory,
	toolRegistry,
	"You are a helpful assistant.",
	50, // 更大的上下文大小
)
```

#### 工作流程示例

```go
// Agent 工作流程示例
func agentWorkflow() {
	agent := createAgent()

	// 第一步：用户输入
	userInput := "Write a function to calculate Fibonacci numbers"

	// 第二步：执行对话
	response, err := agent.Run(userInput)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// 第三步：处理响应
	fmt.Printf("Agent response: %s\n", response.Content)

	// 第四步：继续对话
	followUpInput := "Can you explain how this function works?"
	response, err = agent.Run(followUpInput)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Agent response: %s\n", response.Content)

	// 第五步：清除上下文（如需）
	// agent.ClearContext()
}
```

### 记忆系统

记忆系统负责存储和管理对话历史。

#### 接口定义

```go
type Memory interface {
	Add(msg Message) error
	GetRecent(n int) ([]Message, error)
	Query(query string, n int) ([]Message, error)
	GetAll() ([]Message, error)
	Clear() error
}
```

### 工具系统

工具系统允许 Agent 执行各种操作，如文件操作、网络搜索等。

#### 内置工具

- `filetool`: 提供文件读写、目录列出和文件搜索功能
- `search`: 提供网络搜索功能
- `crawler`: 提供网页爬取功能
- `time`: 提供时间相关功能
- `web`: 提供网页访问功能

### LLM 客户端

LLM 客户端负责与语言模型进行交互。

#### 支持的后端

- OpenAI: 支持 GPT 系列模型
- Ollama: 支持本地部署的模型

## API 参考

### 核心类型

#### Message

```go
type Message struct {
	Role    Role   // 消息角色：system, user, assistant, tool
	Content string // 消息内容
	ToolCalls  []ToolCall // 工具调用
	ToolCallID string     // 工具调用 ID
	Images []ImageContent // 图像内容
	Audio  *AudioContent  // 音频内容
	Video  *VideoContent  // 视频内容
	Metadata  map[string]any // 元数据
	Timestamp int64          // 时间戳
}
```

#### Tool

```go
type Tool interface {
	Name() string
	Description() string
	Schema() ToolSchema
	Execute(input ToolInput) (ToolOutput, error)
}
```

### 配置选项

配置文件 `config/config.yaml` 支持以下选项：

```yaml
# 代理配置
agent:
  system_prompt: "You are a helpful assistant."
  max_context_size: 10

# LLM 配置
llm:
  provider: "openai" # 或 "ollama"
  openai:
    api_key: "your-openai-api-key"
    model: "gpt-4"
  ollama:
    base_url: "http://localhost:11434"
    model: "llama3"

# 工具配置
tools:
  enabled: true
  builtin:
    file: true
    search: true
    crawler: true
    time: true
    web: true
```

## 示例

### 1. 连续对话示例

```go
// 连续对话示例
package main

import (
	"github.com/DaWesen/Pandora/pkg/chat"
	"github.com/DaWesen/Pandora/pkg/core"
	"github.com/DaWesen/Pandora/pkg/core/memory"
	"github.com/DaWesen/Pandora/pkg/core/tool"
	"github.com/DaWesen/Pandora/pkg/core/tool/builtin"
	"github.com/DaWesen/Pandora/pkg/llm/openai"
)

func createAgent() *core.AgentImpl {
	// 创建 LLM 客户端
	llmClient := openai.NewClient("your-openai-api-key")

	// 创建记忆系统
	shortTermMemory := memory.NewShortTermMemory()

	// 创建工具注册表
	toolRegistry := tool.NewRegistry()

	// 注册内置工具
	fileTool := builtin.NewFileTool()
	toolRegistry.Register(fileTool)

	// 创建 Agent
	agent := core.NewAgent(
		llmClient,
		shortTermMemory,
		toolRegistry,
		"You are a helpful assistant.",
		10,
	)

	return agent
}

func main() {
	// 创建 Agent
	agent := createAgent()

	// 启动连续对话循环
	// 可选参数：prompt 和 exitCommand
	// 这里使用默认值
	chat.ChatLoop(agent, "", "")

	// 或者使用流式对话
	// chat.StreamingChatLoop(agent, "", "")
}
```

### 2. 基本对话示例

```go
// 基本对话示例
func basicConversation() {
	agent := createAgent()

	// 第一次对话
	response, err := agent.Run("Hello, what's your name?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Agent: %s\n", response.Content)

	// 继续对话（上下文保持）
	response, err = agent.Run("What can you help me with?")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Agent: %s\n", response.Content)
}
```

### 3. 工具使用示例

```go
// 工具使用示例
func toolUsageExample() {
	agent := createAgent()

	// 请求 Agent 读取文件
	response, err := agent.Run("Read the README.md file and summarize its content")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Agent: %s\n", response.Content)

	// 请求 Agent 搜索信息
	response, err = agent.Run("Search for the latest Go language features")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Agent: %s\n", response.Content)
}
```

### 4. 流式对话示例

```go
// 流式对话示例
func streamingExample() {
	agent := createAgent()

	// 执行流式对话
	fmt.Println("User: Write a short story about a robot learning to paint")
	fmt.Println("Agent:")

	msgChan, errChan := agent.StreamRun("Write a short story about a robot learning to paint")

	// 处理流式响应
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				fmt.Println()
				fmt.Println("Story complete!")
				return
			}
			fmt.Printf("%s", msg.Content)
		case err, ok := <-errChan:
			if !ok {
				return
			}
			fmt.Printf("Error: %v\n", err)
			return
		}
	}
}
```

## 扩展 Agent 框架

### 1. 创建自定义工具

```go
// 自定义工具示例
package mytools

import (
	"encoding/json"
	"fmt"

	"github.com/DaWesen/Pandora/pkg/core"
)

type CalculatorTool struct{}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

func (t *CalculatorTool) Name() string {
	return "calculator"
}

func (t *CalculatorTool) Description() string {
	return "Performs arithmetic calculations"
}

func (t *CalculatorTool) Schema() core.ToolSchema {
	return core.ToolSchema{
		Type: "object",
		Properties: map[string]core.Property{
			"operation": {
				Type:        "string",
				Description: "Operation: add, subtract, multiply, divide",
				Enum:        []string{"add", "subtract", "multiply", "divide"},
			},
			"num1": {
				Type:        "number",
				Description: "First number",
			},
			"num2": {
				Type:        "number",
				Description: "Second number",
			},
		},
		Required: []string{"operation", "num1", "num2"},
	}
}

func (t *CalculatorTool) Execute(input core.ToolInput) (core.ToolOutput, error) {
	var params struct {
		Operation string  `json:"operation"`
		Num1      float64 `json:"num1"`
		Num2      float64 `json:"num2"`
	}

	if err := json.Unmarshal(input.Arguments, &params); err != nil {
		return core.ToolOutput{}, err
	}

	var result float64
	switch params.Operation {
	case "add":
		result = params.Num1 + params.Num2
	case "subtract":
		result = params.Num1 - params.Num2
	case "multiply":
		result = params.Num1 * params.Num2
	case "divide":
		if params.Num2 == 0 {
			return core.ToolOutput{}, fmt.Errorf("division by zero")
		}
		result = params.Num1 / params.Num2
	default:
		return core.ToolOutput{}, fmt.Errorf("unknown operation: %s", params.Operation)
	}

	return core.ToolOutput{
		Content: fmt.Sprintf("%f %s %f = %f", params.Num1, params.Operation, params.Num2, result),
		Data: map[string]any{
			"result": result,
		},
	}, nil
}
```

### 2. 扩展记忆系统

```go
// 自定义记忆系统示例
package mymemory

import (
	"github.com/DaWesen/Pandora/pkg/core"
)

type EnhancedMemory struct {
	shortTerm []core.Message
	maxSize   int
}

func NewEnhancedMemory(maxSize int) *EnhancedMemory {
	return &EnhancedMemory{
		shortTerm: make([]core.Message, 0, maxSize),
		maxSize:   maxSize,
	}
}

func (m *EnhancedMemory) Add(msg core.Message) error {
	m.shortTerm = append(m.shortTerm, msg)
	// 保持记忆大小在限制范围内
	if len(m.shortTerm) > m.maxSize {
		m.shortTerm = m.shortTerm[len(m.shortTerm)-m.maxSize:]
	}
	return nil
}

func (m *EnhancedMemory) GetRecent(n int) ([]core.Message, error) {
	if n > len(m.shortTerm) {
		n = len(m.shortTerm)
	}
	return m.shortTerm[len(m.shortTerm)-n:], nil
}

func (m *EnhancedMemory) Query(query string, n int) ([]core.Message, error) {
	// 简单实现：返回最近的消息
	return m.GetRecent(n)
}

func (m *EnhancedMemory) GetAll() ([]core.Message, error) {
	return m.shortTerm, nil
}

func (m *EnhancedMemory) Clear() error {
	m.shortTerm = make([]core.Message, 0, m.maxSize)
	return nil
}
```

### 3. 实现自定义 LLM 客户端

```go
// 自定义 LLM 客户端示例
package myllm

import (
	"github.com/DaWesen/Pandora/pkg/core"
)

type CustomLLMClient struct {
	apiKey string
	model  string
}

func NewCustomLLMClient(apiKey, model string) *CustomLLMClient {
	return &CustomLLMClient{
		apiKey: apiKey,
		model:  model,
	}
}

func (c *CustomLLMClient) Chat(messages []core.Message, tools []core.Tool) (core.Message, error) {
	// 实现与自定义 LLM 的交互
	// ...
	return core.Message{
		Role:    core.RoleAssistant,
		Content: "This is a response from custom LLM",
	}, nil
}

func (c *CustomLLMClient) ChatStream(messages []core.Message, tools []core.Tool) (<-chan core.Message, <-chan error) {
	// 实现流式响应
	msgChan := make(chan core.Message)
	errChan := make(chan error)

	// 启动goroutine处理流式响应
	go func() {
		defer close(msgChan)
		defer close(errChan)

		// 模拟流式响应
		msgChan <- core.Message{
			Role:    core.RoleAssistant,
			Content: "Streaming response from custom LLM",
		}
	}()

	return msgChan, errChan
}
```

## 最佳实践

### 1. 系统提示设计

- **明确角色**：为 Agent 定义清晰的角色和职责
- **设定边界**：明确 Agent 能做什么和不能做什么
- **提供指导**：给出处理特定任务的指导方针
- **保持简洁**：系统提示应简洁明了，避免过长

### 2. 工具使用

- **合理注册**：只注册必要的工具，避免工具过多导致混乱
- **工具描述**：为工具提供清晰、准确的描述，帮助 LLM 理解工具的用途
- **错误处理**：实现健壮的错误处理，确保工具执行失败时能优雅处理

### 3. 记忆管理

- **上下文大小**：根据 LLM 的能力设置适当的上下文大小
- **记忆清理**：在适当的时候清理上下文，避免内存占用过高
- **记忆优化**：考虑实现更高级的记忆管理策略，如记忆摘要和重要信息提取

### 4. 性能优化

- **并发处理**：利用 Go 的并发特性处理多个对话
- **缓存机制**：对频繁使用的工具结果进行缓存
- **批量操作**：使用批量注册等方法减少重复操作

## 常见问题

### 1. Agent 不执行工具调用

**可能原因**：
- 系统提示中没有明确指示 Agent 可以使用工具
- 工具描述不够清晰，LLM 无法理解工具的用途
- LLM 模型不支持工具调用

**解决方案**：
- 在系统提示中明确指示 Agent 可以使用工具
- 为工具提供更详细、准确的描述
- 使用支持工具调用的 LLM 模型

### 2. 记忆系统占用内存过高

**可能原因**：
- 上下文大小设置过大
- 对话历史过长

**解决方案**：
- 调整上下文大小
- 定期清理上下文
- 实现更高效的记忆管理策略

### 3. 流式响应不流畅

**可能原因**：
- LLM 响应速度慢
- 网络延迟高

**解决方案**：
- 使用更快的 LLM 模型
- 优化网络连接
- 实现响应缓存机制

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT
