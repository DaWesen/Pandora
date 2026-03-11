package core

import (
	"time"
)

type AgentImpl struct {
	llm            LLMClient
	memory         Memory
	toolRegistry   ToolRegistry
	systemPrompt   string
	maxContextSize int
}

// 创建新的agent实例
func NewAgent(llm LLMClient, memory Memory, toolRegistry ToolRegistry,
	systemPrompt string, maxContextSize int) *AgentImpl {
	return &AgentImpl{
		llm:            llm,
		memory:         memory,
		toolRegistry:   toolRegistry,
		systemPrompt:   systemPrompt,
		maxContextSize: maxContextSize,
	}
}

// Run执行一次完整的对话流程
func (a *AgentImpl) Run(input string) (*Message, error) {
	userMsg := Message{
		Role:      RoleUser,
		Content:   input,
		Timestamp: time.Now().Unix(),
	}
	if err := a.memory.Add(userMsg); err != nil {
		return nil, err
	}
	systemMsg := Message{
		Role:    RoleSystem,
		Content: a.systemPrompt,
	}
	recentMsgs, err := a.memory.GetRecent(a.maxContextSize)
	if err != nil {
		return nil, err
	}
	messages := []Message{systemMsg}
	messages = append(messages, recentMsgs...)
	tools := a.toolRegistry.List()
	response, err := a.llm.Chat(messages, tools)
	if err != nil {
		return nil, err
	}
	if len(response.ToolCalls) > 0 {
		for _, toolCall := range response.ToolCalls {
			toolInput := ToolInput{
				Arguments: toolCall.Function.Arguments,
			}
			toolOutput, err := a.toolRegistry.Execute(toolCall.Function.Name, toolInput)
			if err != nil {
				toolResponse := Message{
					Role:       RoleTool,
					Content:    "Error executing tool: " + err.Error(),
					ToolCallID: toolCall.ID,
				}
				if err := a.memory.Add(toolResponse); err != nil {
					return nil, err
				}
			} else {
				toolResponse := Message{
					Role:       RoleTool,
					Content:    toolOutput.Content,
					ToolCallID: toolCall.ID,
				}
				if err := a.memory.Add(toolResponse); err != nil {
					return nil, err
				}
				messages = append(messages, response)
				messages = append(messages, toolResponse)
				response, err = a.llm.Chat(messages, tools)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	if err := a.memory.Add(response); err != nil {
		return nil, err
	}
	return &response, nil
}

// AddMessage 直接添加消息到记忆中
func (a *AgentImpl) AddMessage(msg Message) error {
	return a.memory.Add(msg)
}

// ClearContext 清除所有记忆
func (a *AgentImpl) ClearContext() error {
	return a.memory.Clear()
}

// RegisterTool 注册工具到工具库
func (a *AgentImpl) RegisterTool(tool Tool) error {
	return a.toolRegistry.Register(tool)
}

// SystemPrompt 返回系统提示
func (a *AgentImpl) SystemPrompt() string {
	return a.systemPrompt
}

// StreamRun 执行流式对话
func (a *AgentImpl) StreamRun(input string) (<-chan Message, <-chan error) {
	userMsg := Message{
		Role:      RoleUser,
		Content:   input,
		Timestamp: time.Now().Unix(),
	}
	if err := a.memory.Add(userMsg); err != nil {
		errChan := make(chan error)
		close(errChan)
		return nil, errChan
	}

	systemMsg := Message{
		Role:    RoleSystem,
		Content: a.systemPrompt,
	}

	recentMsgs, err := a.memory.GetRecent(a.maxContextSize)
	if err != nil {
		errChan := make(chan error)
		close(errChan)
		return nil, errChan
	}

	messages := []Message{systemMsg}
	messages = append(messages, recentMsgs...)
	tools := a.toolRegistry.List()

	// 调用流式聊天
	return a.llm.ChatStream(messages, tools)
}
