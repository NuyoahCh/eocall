package mem

import (
	"github.com/cloudwego/eino/schema"
	"sync"
)

// SimpleMemoryMap 创建 map kv 键值对存储
var SimpleMemoryMap = make(map[string]*SimpleMemory)

// 互斥锁
var mu sync.Mutex

// GetSimpleMemory 获取内存信息
func GetSimpleMemory(id string) *SimpleMemory {
	mu.Lock()
	defer mu.Unlock()
	// 如果存在就返回，不存在就创建
	if mem, ok := SimpleMemoryMap[id]; ok {
		return mem
	} else {
		newMem := &SimpleMemory{
			ID:            id,
			Messages:      []*schema.Message{},
			MaxWindowSize: 6,
		}
		SimpleMemoryMap[id] = newMem
		return newMem
	}
}

// SimpleMemory 初始化内存参数
type SimpleMemory struct {
	ID            string            `json:"id"`
	Messages      []*schema.Message `json:"messages"`
	MaxWindowSize int
	mu            sync.Mutex
}

// SetMessages 设置消息
func (c *SimpleMemory) SetMessages(msg *schema.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Messages = append(c.Messages, msg)
	if len(c.Messages) > c.MaxWindowSize {
		// 确保成对丢弃消息，保持对话配对关系
		// 计算需要丢弃的消息数量（必须是偶数）
		excess := len(c.Messages) - c.MaxWindowSize
		if excess%2 != 0 {
			excess++ // 确保丢弃偶数条消息
		}
		// 丢弃前面的消息，保持对话配对
		c.Messages = c.Messages[excess:]
	}
}

// GetMessages 获取消息内容
func (c *SimpleMemory) GetMessages() []*schema.Message {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Messages
}
