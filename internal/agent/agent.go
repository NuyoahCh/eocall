package agent

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/cloudwego/eino/schema"

	"github.com/NuyoahCh/eocall/internal/agent/executor"
	"github.com/NuyoahCh/eocall/internal/agent/planner"
	"github.com/NuyoahCh/eocall/internal/chat/memory"
	"github.com/NuyoahCh/eocall/internal/chat/session"
	"github.com/NuyoahCh/eocall/internal/llm/eino"
	"github.com/NuyoahCh/eocall/internal/rag"
	"github.com/NuyoahCh/eocall/internal/tools/registry"
)

// Agent OnCall 智能体
type Agent struct {
	planner      *planner.Planner
	executor     *executor.Executor
	toolRegistry *registry.Registry
	ragService   *rag.Service
	memory       *memory.SlidingWindowMemory
	sessionMgr   *session.Manager
	llmClient    *eino.Client
}

// Config Agent 配置
type Config struct {
	MaxHistory   int
	SummaryAfter int
}

// NewAgent 创建 Agent
func NewAgent(
	llmClient *eino.Client,
	toolRegistry *registry.Registry,
	ragService *rag.Service,
	sessionMgr *session.Manager,
	cfg *Config,
) *Agent {
	p := planner.NewPlanner(llmClient)
	exec := executor.NewExecutor(toolRegistry, p)

	return &Agent{
		planner:      p,
		executor:     exec,
		toolRegistry: toolRegistry,
		ragService:   ragService,
		sessionMgr:   sessionMgr,
		llmClient:    llmClient,
		memory:       memory.NewSlidingWindowMemory(cfg.MaxHistory, cfg.SummaryAfter, nil),
	}
}

// Request 请求
type Request struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// Response 响应
type Response struct {
	Message string         `json:"message"`
	Plan    *planner.Plan  `json:"plan,omitempty"`
	Sources []rag.Document `json:"sources,omitempty"`
}

// Chat 对话
func (a *Agent) Chat(ctx context.Context, req *Request) (*Response, error) {
	// 1. 获取或创建会话
	sess := a.sessionMgr.GetOrCreate(ctx, req.UserID, req.SessionID)
	sess.AddMessage("user", req.Message)

	// 2. 构建上下文 (滑动窗口)
	summary, recentMsgs, err := a.memory.BuildContext(ctx, sess)
	if err != nil {
		return nil, err
	}

	// 3. RAG 检索相关知识
	var ragContext string
	var sources []rag.Document
	if a.ragService != nil {
		docs, err := a.ragService.Retrieve(ctx, req.Message, 5)
		if err == nil && len(docs) > 0 {
			ragContext = a.ragService.FormatContext(docs)
			sources = docs
		}
	}

	// 4. 构建消息列表
	messages := a.buildMessages(summary, recentMsgs, ragContext, req.Message)

	// 5. 调用 LLM
	resp, err := a.llmClient.GenerateWithMessages(ctx, messages)
	if err != nil {
		return nil, err
	}

	sess.AddMessage("assistant", resp.Content)

	return &Response{
		Message: resp.Content,
		Sources: sources,
	}, nil
}

// ChatStream 流式对话
func (a *Agent) ChatStream(ctx context.Context, req *Request, callback func(chunk string)) error {
	sess := a.sessionMgr.GetOrCreate(ctx, req.UserID, req.SessionID)
	sess.AddMessage("user", req.Message)

	summary, recentMsgs, _ := a.memory.BuildContext(ctx, sess)

	var ragContext string
	if a.ragService != nil {
		docs, _ := a.ragService.Retrieve(ctx, req.Message, 5)
		if len(docs) > 0 {
			ragContext = a.ragService.FormatContext(docs)
		}
	}

	messages := a.buildMessages(summary, recentMsgs, ragContext, req.Message)

	// 流式调用
	stream, err := a.llmClient.GetChatModel().Stream(ctx, messages)
	if err != nil {
		return err
	}
	defer stream.Close()

	var fullResponse strings.Builder
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		fullResponse.WriteString(chunk.Content)
		callback(chunk.Content)
	}

	sess.AddMessage("assistant", fullResponse.String())
	return nil
}

func (a *Agent) buildMessages(summary string, recentMsgs []session.Message, ragContext, userMessage string) []*schema.Message {
	messages := make([]*schema.Message, 0)

	// 系统提示
	systemContent := `你是一个专业的运维 AI 助手，负责帮助用户分析告警、排查故障、解答运维问题。
请根据上下文信息和用户问题，给出专业、准确的回答。`

	if summary != "" {
		systemContent += "\n\n## 历史摘要\n" + summary
	}
	if ragContext != "" {
		systemContent += "\n\n## 相关知识\n" + ragContext
	}

	messages = append(messages, &schema.Message{
		Role:    schema.System,
		Content: systemContent,
	})

	// 历史消息
	for _, msg := range recentMsgs {
		role := schema.User
		if msg.Role == "assistant" {
			role = schema.Assistant
		}
		messages = append(messages, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	// 当前用户消息 (如果不在历史中)
	if len(recentMsgs) == 0 || recentMsgs[len(recentMsgs)-1].Content != userMessage {
		messages = append(messages, &schema.Message{
			Role:    schema.User,
			Content: userMessage,
		})
	}

	return messages
}
