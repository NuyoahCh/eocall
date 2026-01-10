package session

import (
	"context"
	"sync"
	"time"
)

// Message 消息
type Message struct {
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Session 会话
type Session struct {
	ID        string
	UserID    string
	Messages  []Message
	Summary   string // 历史摘要
	CreatedAt time.Time
	UpdatedAt time.Time
	mu        sync.RWMutex
}

// NewSession 创建新会话
func NewSession(id, userID string) *Session {
	now := time.Now()
	return &Session{
		ID:        id,
		UserID:    userID,
		Messages:  make([]Message, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddMessage 添加消息
func (s *Session) AddMessage(role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Messages = append(s.Messages, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	s.UpdatedAt = time.Now()
}

// GetMessages 获取消息
func (s *Session) GetMessages() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Message, len(s.Messages))
	copy(result, s.Messages)
	return result
}

// GetRecentMessages 获取最近 N 条消息
func (s *Session) GetRecentMessages(n int) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.Messages) <= n {
		result := make([]Message, len(s.Messages))
		copy(result, s.Messages)
		return result
	}

	result := make([]Message, n)
	copy(result, s.Messages[len(s.Messages)-n:])
	return result
}

// SetSummary 设置历史摘要
func (s *Session) SetSummary(summary string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Summary = summary
}

// GetSummary 获取历史摘要
func (s *Session) GetSummary() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Summary
}

// MessageCount 消息数量
func (s *Session) MessageCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Messages)
}

// Manager 会话管理器 - 基于 UserID 的会话隔离
type Manager struct {
	sessions map[string]*Session // key: userID:sessionID
	mu       sync.RWMutex
	ttl      time.Duration
}

// NewManager 创建会话管理器
func NewManager(ttl time.Duration) *Manager {
	m := &Manager{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}
	go m.cleanupLoop()
	return m
}

// GetOrCreate 获取或创建会话
func (m *Manager) GetOrCreate(ctx context.Context, userID, sessionID string) *Session {
	key := userID + ":" + sessionID

	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.sessions[key]; ok {
		s.UpdatedAt = time.Now()
		return s
	}

	s := NewSession(sessionID, userID)
	m.sessions[key] = s
	return s
}

// Get 获取会话
func (m *Manager) Get(ctx context.Context, userID, sessionID string) (*Session, bool) {
	key := userID + ":" + sessionID

	m.mu.RLock()
	defer m.mu.RUnlock()

	s, ok := m.sessions[key]
	return s, ok
}

// Delete 删除会话
func (m *Manager) Delete(ctx context.Context, userID, sessionID string) {
	key := userID + ":" + sessionID

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, key)
}

// cleanupLoop 定期清理过期会话
func (m *Manager) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.cleanup()
	}
}

func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, s := range m.sessions {
		if now.Sub(s.UpdatedAt) > m.ttl {
			delete(m.sessions, key)
		}
	}
}
