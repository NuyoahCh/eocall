package planner

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/schema"
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
	GenerateWithMessages(ctx context.Context, messages []*schema.Message) (*schema.Message, error)
	GenerateStream(ctx context.Context, prompt string, callback func(chunk string)) error
}

// Planner Plan-Execute 规划器
type Planner struct {
	llm LLMClient
}

// NewPlanner 创建规划器
func NewPlanner(llm LLMClient) *Planner {
	return &Planner{llm: llm}
}

// CreatePlan 创建执行计划
func (p *Planner) CreatePlan(ctx context.Context, goal string, tools []string, context string) (*Plan, error) {
	messages := []*schema.Message{
		{Role: schema.System, Content: systemPrompt},
		{Role: schema.User, Content: p.buildPlanningPrompt(goal, tools, context)},
	}

	response, err := p.llm.GenerateWithMessages(ctx, messages)
	if err != nil {
		return nil, err
	}

	plan, err := p.parsePlanResponse(response.Content)
	if err != nil {
		// 解析失败，返回简单计划
		return &Plan{
			Goal:   goal,
			Status: "pending",
			Steps: []Step{
				{ID: 1, Description: response.Content, Status: "pending"},
			},
		}, nil
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

	// 检查是否需要修订
	if strings.Contains(response, "no_change") {
		return nil, nil
	}

	return p.parsePlanResponse(response)
}

func (p *Planner) buildPlanningPrompt(goal string, tools []string, context string) string {
	var sb strings.Builder
	sb.WriteString("## 可用工具\n")
	for _, t := range tools {
		sb.WriteString("- ")
		sb.WriteString(t)
		sb.WriteString("\n")
	}

	if context != "" {
		sb.WriteString("\n## 上下文信息\n")
		sb.WriteString(context)
	}

	sb.WriteString("\n## 用户目标\n")
	sb.WriteString(goal)

	sb.WriteString(`

## 输出格式
请以 JSON 格式输出执行计划:
` + "```json" + `
{
  "steps": [
    {"id": 1, "description": "步骤描述", "tool_name": "工具名", "tool_params": {"param": "value"}}
  ]
}
` + "```" + `

如果不需要使用工具，可以省略 tool_name 和 tool_params，直接回答用户问题。`)

	return sb.String()
}

func (p *Planner) buildRevisionPrompt(plan *Plan, stepResult string) string {
	planJSON, _ := json.Marshal(plan)
	return `根据当前执行结果，判断是否需要修订计划。

## 当前计划
` + string(planJSON) + `

## 最新执行结果
` + stepResult + `

如果需要修订计划，请输出新的计划 JSON。如果不需要修订，输出 {"no_change": true}。`
}

func (p *Planner) parsePlanResponse(response string) (*Plan, error) {
	// 尝试提取 JSON
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		jsonStr = response
	}

	var result struct {
		Steps []Step `json:"steps"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}

	for i := range result.Steps {
		if result.Steps[i].Status == "" {
			result.Steps[i].Status = "pending"
		}
	}

	return &Plan{Steps: result.Steps}, nil
}

// extractJSON 从文本中提取 JSON
func extractJSON(text string) string {
	// 尝试匹配 ```json ... ``` 格式
	re := regexp.MustCompile("(?s)```json\\s*(.+?)\\s*```")
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}

	// 尝试匹配 { ... } 格式
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start != -1 && end > start {
		return text[start : end+1]
	}

	return ""
}

const systemPrompt = `你是一个专业的运维 AI Agent，负责分析告警、排查故障、执行运维操作。

你的能力包括：
1. 分析用户描述的问题，理解故障现象
2. 制定合理的排查计划
3. 调用工具查询日志、监控指标、告警信息
4. 根据查询结果进行根因分析
5. 给出解决方案和建议

请根据用户的问题，制定执行计划并逐步完成任务。`
