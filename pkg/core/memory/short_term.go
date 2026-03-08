package memory

import (
	"container/list"
	"strings"
	"sync"

	"github.com/DaWesen/Pandora/pkg/core"
)

// 基于内存的短期记忆
type ShortTerm struct {
	maxSize int
	mu      sync.RWMutex
	items   *list.List
}

// 创建短期记忆实例
func NewShortTerm(maxSize int) *ShortTerm {
	if maxSize <= 0 {
		maxSize = 20
	}
	return &ShortTerm{
		maxSize: maxSize,
		items:   list.New(),
	}
}

func (m *ShortTerm) Add(msg core.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 添加新消息
	m.items.PushBack(msg)
	// 超出最大容量时，移除最旧的消息
	for m.items.Len() > m.maxSize {
		m.items.Remove(m.items.Front())
	}
	return nil
}

// 获取最近的n条消息
func (m *ShortTerm) GetRecent(n int) ([]core.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if n <= 0 {
		return []core.Message{}, nil
	}
	if n > m.items.Len() {
		n = m.items.Len()
	}
	result := make([]core.Message, n)
	count := n - 1 // 从切片末尾开始填充
	for e := m.items.Back(); e != nil && count >= 0; e = e.Prev() {
		if msg, ok := e.Value.(core.Message); ok {
			result[count] = msg
			count--
		}
	}
	return result, nil
}

// 获取所有消息
func (m *ShortTerm) GetAll() ([]core.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]core.Message, 0, m.items.Len())
	for e := m.items.Front(); e != nil; e = e.Next() {
		if msg, ok := e.Value.(core.Message); ok {
			result = append(result, msg)
		}
	}
	return result, nil
}

// 根据查询获取相关消息
func (m *ShortTerm) Query(query string, n int) ([]core.Message, error) {
	// 简单的查询实现：返回包含查询字符串的消息
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []core.Message
	for e := m.items.Front(); e != nil; e = e.Next() {
		if msg, ok := e.Value.(core.Message); ok {
			if strings.Contains(msg.Content, query) {
				result = append(result, msg)
				if n > 0 && len(result) >= n {
					break
				}
			}
		}
	}
	return result, nil
}

func (m *ShortTerm) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items.Init() //清空链表
	return nil
}

func (m *ShortTerm) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.items.Len()
}
