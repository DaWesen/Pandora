package tool

import (
	"fmt"
	"sync"

	"github.com/DaWesen/Pandora/pkg/core"
)

type Registry struct {
	mu    sync.RWMutex
	tools map[string]core.Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]core.Tool),
	}
}

// 注册工具
func (r *Registry) Register(tool core.Tool) error {
	if tool == nil {
		return fmt.Errorf("cannot register nil tool")
	}
	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool with name '%s' already exists", name)
	}
	r.tools[name] = tool
	return nil
}

// 获得指定名称的工具
func (r *Registry) Get(name string) (core.Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, exists := r.tools[name]
	return tool, exists
}

// 列出所有已经注册的工具
func (r *Registry) List() []core.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]core.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		list = append(list, tool)
	}
	return list
}

// 获取所有工具名称
func (r *Registry) ListNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// 获取工具数量
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// 注销工具
func (r *Registry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[name]; exists {
		delete(r.tools, name)
		return true
	}
	return false
}

// 直接执行指定工具
func (r *Registry) Execute(name string, input core.ToolInput,
) (core.ToolOutput, error) {
	tool, exists := r.Get(name)
	if !exists {
		return core.ToolOutput{}, fmt.Errorf("tool with name '%s' not found", name)
	}
	return tool.Execute(input)
}

func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools = make(map[string]core.Tool)
}

// 转化为OpenAI函数调用格式
func (r *Registry) ToOpenAIFunctions() []core.OpenAIFunction {
	r.mu.RLock()
	defer r.mu.RUnlock()
	functions := make([]core.OpenAIFunction, 0, len(r.tools))
	for _, tool := range r.tools {
		schema := tool.Schema()
		function := core.OpenAIFunction{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters: map[string]interface{}{
				"type":       schema.Type,
				"properties": schema.Properties,
				"required":   schema.Required,
			},
		}
		functions = append(functions, function)
	}
	return functions
}

// 转化为Ollama工具格式
func (r *Registry) ToOllamaTools() []core.OllamaTool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tools := make([]core.OllamaTool, 0, len(r.tools))
	for _, tool := range r.tools {
		schema := tool.Schema()
		ollamaTool := core.OllamaTool{
			Type: "function",
			Function: map[string]interface{}{
				"name":        tool.Name(),
				"description": tool.Description(),
				"parameters":  schema.Properties,
			},
		}
		tools = append(tools, ollamaTool)
	}
	return tools
}
