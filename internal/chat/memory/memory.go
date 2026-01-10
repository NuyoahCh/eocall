package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/NuyoahCh/eocall/internal/chat/session"
)

// Summarizer 摘要生成器接口
type Summarizer interface {
	Summarize(ctx context.Context, messages []session.Message) (string, error)
}

// SlidingWindowMemory 滑动窗口记忆管理
// 设计滑动窗口摘要机制，保留最近 50 轮对话并压缩历史
type SlidingWindowMemory struct {
	maxHistory   int        // 最大保留历史轮数
	summaryAfter int        // 多少轮后开始摘要
	summarizer   Summarizer // 摘要生成器
}

// NewSlidingWindowMemory 创建滑动窗口记忆
func NewSlidingWindowMemory(maxHistory, summaryAfter int, summarizer Summarizer) *SlidingWindowMemory {
	return &SlidingWindowMemory{
		maxHistory:   maxHistory,
		summaryAfter: summaryAfter,
		summarizer:   summarizer,
	}
}

// BuildContext 构建上下文
// 返回: 系统提示(含摘要) + 最近消息
func (m *SlidingWindowMemory) BuildContext(ctx context.Context, s *session.Session) (string, []session.Message, error) {
	messages := s.GetMessages()
	summary := s.GetSummary()

	// 消息数量未超过阈值，直接返回
	if len(messages) <= m.maxHistory {
		return summary, messages, nil
	}

	// 需要压缩历史
	// 保留最近 maxHistory 条消息
	recentMessages := messages[len(messages)-m.maxHistory:]

	// 需要摘要的消息
	toSummarize := messages[:len(messages)-m.maxHistory]

	// 生成新摘要
	if m.summarizer != nil && len(toSummarize) >= m.summaryAfter {
		newSummary, err := m.summarizer.Summarize(ctx, toSummarize)
		if err != nil {
			// 摘要失败，使用旧摘要
			return summary, recentMessages, nil
		}

		// 合并摘要
		if summary != "" {
			summary = summary + "\n\n" + newSummary
		} else {
			summary = newSummary
		}
		s.SetSummary(summary)
	}

	return summary, recentMessages, nil
}

// FormatMessages 格式化消息为字符串
func FormatMessages(messages []session.Message) string {
	var sb strings.Builder
	for _, msg := range messages {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content))
	}
	return sb.String()
}

// EstimateTokens 估算 Token 数量 (简单估算)
func EstimateTokens(text string) int {
	// 简单估算: 中文约 1.5 字符/token, 英文约 4 字符/token
	// 这里使用保守估算
	return len(text) / 2
}
