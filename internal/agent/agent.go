package agent

import (
	"context"

	"github.com/NuyoahCh/eocall/internal/agent/executor"
	"github.com/NuyoahCh/eocall/internal/agent/planner"
	"github.com/NuyoahCh/eocall/internal/chat/memory"
	"github.com/NuyoahCh/eocall/internal/chat/session"
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
}

// Config Agent 配置
type Config struct {
	MaxHistory   int
	SummaryAfter int
}

// NewAgent 创建 Agent
func NewAgent(
	p *planner.Planner,
	toolRegistry *registry.Registry,
	ragService *rag.Service,
	sessionMgr *session.Manager,
	cfg *Config,
) *Agent {
	exec := executor.NewExecutor(toolRegistry, p)

	return &Agent{
		planner:      p,
		executor:     exec,
		toolRegistry: toolRegistry,
		ragService:   ragService,
		sessionMgr:   sessionMgr,
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

	// 4. 构建完整上下文
	fullContext := buildFullContext(summary, recentMsgs, ragContext)

	// 5. 创建执行计划
	toolNames := getToolNames(a.toolRegistry.List())
	plan, err := a.planner.CreatePlan(ctx, req.Message, toolNames, fullContext)
	if err != nil {
		return nil, err
	}

	// 6. 执行计划
	if err := a.executor.ExecutePlan(ctx, plan); err != nil {
		return nil, err
	}

	// 7. 生成响应
	response := a.generateResponse(plan)
	sess.AddMessage("assistant", response)

	return &Response{
		Message: response,
		Plan:    plan,
		Sources: sources,
	}, nil
}

// ChatStream 流式对话
func (a *Agent) ChatStream(ctx context.Context, req *Request, callback func(chunk string)) error {
	// 类似 Chat，但使用流式输出
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

	fullContext := buildFullContext(summary, recentMsgs, ragContext)
	toolNames := getToolNames(a.toolRegistry.List())

	plan, err := a.planner.CreatePlan(ctx, req.Message, toolNames, fullContext)
	if err != nil {
		return err
	}

	// 流式执行
	var fullResponse string
	err = a.executor.ExecuteStepByStep(ctx, plan, func(step *planner.Step, result string) {
		chunk := formatStepResult(step, result)
		fullResponse += chunk
		callback(chunk)
	})

	if err != nil {
		return err
	}

	sess.AddMessage("assistant", fullResponse)
	return nil
}

func buildFullContext(summary string, msgs []session.Message, ragContext string) string {
	ctx := ""
	if summary != "" {
		ctx += "## 历史摘要\n" + summary + "\n\n"
	}
	if len(msgs) > 0 {
		ctx += "## 最近对话\n" + memory.FormatMessages(msgs) + "\n\n"
	}
	if ragContext != "" {
		ctx += "## 相关知识\n" + ragContext + "\n\n"
	}
	return ctx
}

func getToolNames(defs []registry.ToolDefinition) []string {
	names := make([]string, len(defs))
	for i, d := range defs {
		names[i] = d.Name + ": " + d.Description
	}
	return names
}

func (a *Agent) generateResponse(plan *planner.Plan) string {
	// 根据执行结果生成响应
	response := ""
	for _, step := range plan.Steps {
		if step.Result != "" {
			response += step.Result + "\n"
		}
	}
	if response == "" {
		response = "任务已完成"
	}
	return response
}

func formatStepResult(step *planner.Step, result string) string {
	if step.Status == "failed" {
		return "❌ " + step.Description + ": " + step.Error + "\n"
	}
	return "✅ " + step.Description + "\n" + result + "\n"
}
