package executor

import (
	"context"
	"fmt"

	"github.com/NuyoahCh/eocall/internal/agent/planner"
	"github.com/NuyoahCh/eocall/internal/tools/registry"
	"github.com/NuyoahCh/eocall/pkg/logger"
)

// Executor 计划执行器
type Executor struct {
	toolRegistry *registry.Registry
	planner      *planner.Planner
}

// NewExecutor 创建执行器
func NewExecutor(toolRegistry *registry.Registry, p *planner.Planner) *Executor {
	return &Executor{
		toolRegistry: toolRegistry,
		planner:      p,
	}
}

// ExecutePlan 执行计划
func (e *Executor) ExecutePlan(ctx context.Context, plan *planner.Plan) error {
	plan.Status = "running"

	for i := range plan.Steps {
		step := &plan.Steps[i]

		if step.Status == "skipped" {
			continue
		}

		step.Status = "running"
		logger.Info("executing step", "step_id", step.ID, "description", step.Description)

		if step.ToolName != "" {
			result, err := e.executeStep(ctx, step)
			if err != nil {
				step.Status = "failed"
				step.Error = err.Error()
				plan.Status = "failed"
				return fmt.Errorf("step %d failed: %w", step.ID, err)
			}
			step.Result = result
		}

		step.Status = "completed"

		// 检查是否需要修订计划
		if e.planner != nil && step.Result != "" {
			revisedPlan, err := e.planner.RevisePlan(ctx, plan, step.Result)
			if err == nil && revisedPlan != nil && len(revisedPlan.Steps) > 0 {
				// 更新后续步骤
				e.updateRemainingSteps(plan, revisedPlan, i+1)
			}
		}
	}

	plan.Status = "completed"
	return nil
}

// executeStep 执行单个步骤
func (e *Executor) executeStep(ctx context.Context, step *planner.Step) (string, error) {
	result, err := e.toolRegistry.Execute(ctx, step.ToolName, step.ToolParams)
	if err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("tool execution failed: %s", result.Error)
	}

	return fmt.Sprintf("%v", result.Data), nil
}

// updateRemainingSteps 更新剩余步骤
func (e *Executor) updateRemainingSteps(plan *planner.Plan, revisedPlan *planner.Plan, fromIndex int) {
	// 简单实现: 替换后续步骤
	if fromIndex < len(plan.Steps) {
		plan.Steps = append(plan.Steps[:fromIndex], revisedPlan.Steps...)
	}
}

// ExecuteStepByStep 逐步执行 (支持流式输出)
func (e *Executor) ExecuteStepByStep(ctx context.Context, plan *planner.Plan, callback func(step *planner.Step, result string)) error {
	plan.Status = "running"

	for i := range plan.Steps {
		step := &plan.Steps[i]

		if step.Status == "skipped" {
			continue
		}

		step.Status = "running"

		if step.ToolName != "" {
			result, err := e.executeStep(ctx, step)
			if err != nil {
				step.Status = "failed"
				step.Error = err.Error()
				callback(step, "")
				plan.Status = "failed"
				return err
			}
			step.Result = result
			step.Status = "completed"
			callback(step, result)
		} else {
			step.Status = "completed"
			callback(step, step.Description)
		}
	}

	plan.Status = "completed"
	return nil
}
