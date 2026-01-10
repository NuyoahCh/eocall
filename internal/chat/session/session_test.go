package session

import (
	"context"
	"testing"
	"time"
)

func TestSession_AddMessage(t *testing.T) {
	s := NewSession("sess-1", "user-1")

	s.AddMessage("user", "hello")
	s.AddMessage("assistant", "hi there")

	msgs := s.GetMessages()
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}

	if msgs[0].Role != "user" || msgs[0].Content != "hello" {
		t.Errorf("unexpected first message: %+v", msgs[0])
	}
}

func TestSession_GetRecentMessages(t *testing.T) {
	s := NewSession("sess-1", "user-1")

	for i := 0; i < 10; i++ {
		s.AddMessage("user", "message")
	}

	recent := s.GetRecentMessages(5)
	if len(recent) != 5 {
		t.Errorf("expected 5 messages, got %d", len(recent))
	}
}

func TestSession_Summary(t *testing.T) {
	s := NewSession("sess-1", "user-1")

	s.SetSummary("this is a summary")
	if s.GetSummary() != "this is a summary" {
		t.Errorf("unexpected summary: %s", s.GetSummary())
	}
}

func TestManager_GetOrCreate(t *testing.T) {
	m := NewManager(30 * time.Minute)

	ctx := context.Background()

	// 创建新会话
	s1 := m.GetOrCreate(ctx, "user-1", "sess-1")
	if s1 == nil {
		t.Fatal("expected session, got nil")
	}

	// 获取同一会话
	s2 := m.GetOrCreate(ctx, "user-1", "sess-1")
	if s1 != s2 {
		t.Error("expected same session instance")
	}

	// 不同用户的会话应该隔离
	s3 := m.GetOrCreate(ctx, "user-2", "sess-1")
	if s1 == s3 {
		t.Error("expected different session for different user")
	}
}

func TestManager_SessionIsolation(t *testing.T) {
	m := NewManager(30 * time.Minute)
	ctx := context.Background()

	// 用户1的会话
	s1 := m.GetOrCreate(ctx, "user-1", "sess-1")
	s1.AddMessage("user", "user1 message")

	// 用户2的会话
	s2 := m.GetOrCreate(ctx, "user-2", "sess-1")
	s2.AddMessage("user", "user2 message")

	// 验证隔离
	if len(s1.GetMessages()) != 1 || s1.GetMessages()[0].Content != "user1 message" {
		t.Error("user1 session corrupted")
	}

	if len(s2.GetMessages()) != 1 || s2.GetMessages()[0].Content != "user2 message" {
		t.Error("user2 session corrupted")
	}
}

func TestManager_Delete(t *testing.T) {
	m := NewManager(30 * time.Minute)
	ctx := context.Background()

	m.GetOrCreate(ctx, "user-1", "sess-1")
	m.Delete(ctx, "user-1", "sess-1")

	_, ok := m.Get(ctx, "user-1", "sess-1")
	if ok {
		t.Error("session should be deleted")
	}
}
