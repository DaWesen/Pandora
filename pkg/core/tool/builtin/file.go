package builtin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DaWesen/Pandora/pkg/core"
)

type FileTool struct{}

func NewFileTool() *FileTool {
	return &FileTool{}
}

func (t *FileTool) Name() string {
	return "filetool"
}

func (t *FileTool) Description() string {
	return "Provides file reading and writing capabilities."
}

// 返回工具参数
func (t *FileTool) Schema() core.ToolSchema {
	return core.ToolSchema{
		Type: "object",
		Properties: map[string]core.Property{
			"operation": {
				Type:        "string",
				Description: "操作类型：'read'(读取), 'write'(写入), 'list'(列出), 'search'(查找)",
				Enum:        []string{"read", "write", "list", "search"},
			},
			"path": {
				Type:        "string",
				Description: "文件路径或搜索目录",
			},
			"content": {
				Type:        "string",
				Description: "写入内容，仅在操作为 'write' 时使用",
			},
			"overwrite": {
				Type:        "boolean",
				Description: "是否覆盖文件，仅在操作为 'write' 时使用，默认为 false",
			},
			"pattern": {
				Type:        "string",
				Description: "搜索模式，仅在操作为 'search' 时使用，如 '*.txt'",
			},
			"recursive": {
				Type:        "boolean",
				Description: "是否递归搜索，仅在操作为 'search' 时使用，默认为 true",
			},
		},
		Required: []string{"operation", "path"},
	}
}

// 执行文件操作
func (t *FileTool) Execute(input core.ToolInput) (core.ToolOutput, error) {
	var params struct {
		Operation string `json:"operation"`
		Path      string `json:"path"`
		Content   string `json:"content,omitempty"`
		Overwrite bool   `json:"overwrite,omitempty"`
		Pattern   string `json:"pattern,omitempty"`
		Recursive bool   `json:"recursive,omitempty"`
	}
	if err := json.Unmarshal(input.Arguments, &params); err != nil {
		return core.ToolOutput{}, err
	}
	switch params.Operation {
	case "read":
		return t.readFile(params.Path)
	case "write":
		return t.writeFile(params.Path, params.Content, params.Overwrite)
	case "list":
		return t.listDirectory(params.Path)
	case "search":
		return t.searchFiles(params.Path, params.Pattern, params.Recursive)
	default:
		return core.ToolOutput{}, fmt.Errorf("不支持的操作类型: %s", params.Operation)
	}
}

// 读取文件内容
func (t *FileTool) readFile(path string) (core.ToolOutput, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return core.ToolOutput{}, fmt.Errorf("读取文件失败:%v", err)

	}
	return core.ToolOutput{
		Content: string(content),
		Data: map[string]any{
			"path":    path,
			"content": string(content),
			"size":    len(content),
		},
	}, nil
}

// 写入文件内容
func (t *FileTool) writeFile(path, content string, overwrite bool) (core.ToolOutput, error) {
	if _, err := os.Stat(path); err == nil && !overwrite {
		return core.ToolOutput{}, fmt.Errorf("文件已存在且,令overwrite为true以覆盖")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return core.ToolOutput{}, fmt.Errorf("创建目录失败:%v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return core.ToolOutput{}, fmt.Errorf("写入文件失败:%v", err)
	}
	return core.ToolOutput{
		Content: fmt.Sprintf("文件写入成功：%s", path),
		Data: map[string]any{
			"path":      path,
			"size":      len(content),
			"overwrite": overwrite,
		},
	}, nil
}

// 列出目录内容
func (t *FileTool) listDirectory(path string) (core.ToolOutput, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return core.ToolOutput{}, fmt.Errorf("读取目录失败: %v", err)
	}

	var items []map[string]any
	for _, entry := range entries {
		entryType := "file"
		if entry.IsDir() {
			entryType = "directory"
		}

		// 获取文件详细信息
		info, err := entry.Info()
		var size int64 = 0
		var modTime string = ""
		if err == nil {
			size = info.Size()
			modTime = info.ModTime().Format("2006-01-02 15:04:05")
		}

		items = append(items, map[string]any{
			"name":    entry.Name(),
			"type":    entryType,
			"size":    size,
			"modTime": modTime,
		})
	}

	return core.ToolOutput{
		Content: fmt.Sprintf("目录 %s 包含 %d 个项目", path, len(items)),
		Data: map[string]any{
			"path":  path,
			"items": items,
			"count": len(items),
		},
	}, nil
}

// 查找文件
func (t *FileTool) searchFiles(dir, pattern string, recursive bool) (core.ToolOutput, error) {
	if pattern == "" {
		pattern = "*"
	}

	var matches []string
	var err error

	if recursive {
		// 递归搜索
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				matched, err := filepath.Match(pattern, filepath.Base(path))
				if err != nil {
					return err
				}
				if matched {
					matches = append(matches, path)
				}
			}
			return nil
		})
	} else {
		// 非递归搜索
		searchPath := filepath.Join(dir, pattern)
		matches, err = filepath.Glob(searchPath)
	}

	if err != nil {
		return core.ToolOutput{}, fmt.Errorf("搜索文件失败: %v", err)
	}

	// 收集文件信息
	var items []map[string]any
	for _, path := range matches {
		info, err := os.Stat(path)
		var size int64 = 0
		var modTime string = ""
		if err == nil {
			size = info.Size()
			modTime = info.ModTime().Format("2006-01-02 15:04:05")
		}
		items = append(items, map[string]any{
			"path":    path,
			"size":    size,
			"modTime": modTime,
		})
	}

	return core.ToolOutput{
		Content: fmt.Sprintf("在 %s 中找到 %d 个匹配文件", dir, len(items)),
		Data: map[string]any{
			"directory": dir,
			"pattern":   pattern,
			"recursive": recursive,
			"items":     items,
			"count":     len(items),
		},
	}, nil
}
