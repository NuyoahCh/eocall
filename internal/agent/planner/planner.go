package planner

import (
	"context"
)

// Plan 执行计划
type Plan struct {
	Goal   string `json:"goal"`
	Steps  []Step `json:"steps"`
	Status string `json:"status"` // pending, running, completed, failed
}

// Step 执行步骤
type Step struct {
	ID          int                    `json:"id"`
	Description string                 `json:"description"`
	ToolName    string                 `json:"tool_name,omitempty"`
	ToolParams  map[string]interface{} `json:"tool_params,omitempty"`
	Status      string                 `json:"status"` // pending, running, completed, failed, skipped
	Result      string                 `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// LLMClient LLM 客户端接口
type LLMClient interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateStream(ctx context.Context, prompt string, callback func(chunk string)) error
}

// Planner Plan-Execute 规划器
type Planner struct {
	llm            LLMClient
	systemPrompt   string
	planningPrompt string
}

// NewPlanner 创建规划器
func NewPlanner(llm LLMClient) *Planner {
	return &Planner{
		llm:            llm,
		systemPrompt:   defaultSystemPrompt,
		planningPrompt: defaultPlanningPrompt,
	}
}

// CreatePlan 创建执行计划
func (p *Planner) CreatePlan(ctx context.Context, goal string, tools []string, context string) (*Plan, error) {
	prompt := p.buildPlanningPrompt(goal, tools, context)

	response, err := p.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	plan, err := p.parsePlanResponse(response)
	if err != nil {
		return nil, err
	}

	plan.Goal = goal
	plan.Status = "pending"
	return plan, nil
}

// RevisePlan 根据执行结果修订计划
func (p *Planner) RevisePlan(ctx context.Context, plan *Plan, stepResult string) (*Plan, error) {
	prompt := p.buildRevisionPrompt(plan, stepResult)

	response, err := p.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return p.parsePlanResponse(response)
}

func (p *Planner) buildPlanningPrompt(goal string, tools []string, context string) string {
	return `你是一个运维专家 AI Agent，需要根据用户的问题制定执行计划。

## 可用工具
` + formatTools(tools) + `

## 上下文信息
` + context + `

## 用户目标
` + goal + `

## 输出格式
请以 JSON 格式输出执行计划:
{
  "steps": [
    {"id": 1, "description": "步骤描述", "tool_name": "工具名", "tool_params": {"param": "value"}},
    ...
  ]
}

如果不需要使用工具，可以省略 tool_name 和 tool_params。
请分析问题并制定合理的执行计划。`
}

func (p *Planner) buildRevisionPrompt(plan *Plan, stepResult string) string {
	return `根据当前执行结果，判断是否需要修订计划。

## 当前计划
` + formatPlan(plan) + `

## 最新执行结果
` + stepResult + `

如果需要修订计划，请输出新的计划 JSON。如果不需要修订，输出 {"no_change": true}。`
}

func (p *Planner) parsePlanResponse(response string) (*Plan, error) {
	// TODO: 实现 JSON 解析逻辑
	// 这里需要从 LLM 响应中提取 JSON 并解析
	return &Plan{Steps: []Step{}}, nil
}

func formatTools(tools []string) string {
	result := ""
	for _, t := range tools {
		result += "- " + t + "\n"
	}
	return result
}

func formatPlan(plan *Plan) string {
	result := "目标: " + plan.Goal + "\n步骤:\n"
	for _, s := range plan.Steps {
		result += "  " + s.Description + " [" + s.Status + "]\n"
	}
	return result
}

const defaultSystemPrompt = `你是一个专业的运维 AI Agent，负责分析告警、排查故障、执行运维操作。
你需要根据用户的问题，制定合理的执行计划，并调用相应的工具完成任务。`

const defaultPlanningPrompt = `请分析用户的问题，制定执行计划。`
