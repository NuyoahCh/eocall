package registry

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/NuyoahCh/eocall/pkg/errors"
)

// ToolParameter 工具参数定义
type ToolParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ToolDefinition 工具定义 - 标准化工具协议
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  []ToolParameter `json:"parameters"`
}

// ToolResult 工具执行结果
type ToolResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Tool 工具接口
type Tool interface {
	// Definition 返回工具定义
	Definition() ToolDefinition
	// Execute 执行工具
	Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error)
}

// Registry 工具注册中心
type Registry struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewRegistry 创建工具注册中心
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (r *Registry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	def := tool.Definition()
	r.tools[def.Name] = tool
}

// Get 获取工具
func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, ok := r.tools[name]
	if !ok {
		return nil, errors.ErrToolNotFound
	}
	return tool, nil
}

// List 列出所有工具
func (r *Registry) List() []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	defs := make([]ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		defs = append(defs, tool.Definition())
	}
	return defs
}

// Execute 执行工具
func (r *Registry) Execute(ctx context.Context, name string, params map[string]interface{}) (*ToolResult, error) {
	tool, err := r.Get(name)
	if err != nil {
		return nil, err
	}
	return tool.Execute(ctx, params)
}

// ToJSON 将工具定义转为 JSON (供 LLM 使用)
func (r *Registry) ToJSON() (string, error) {
	defs := r.List()
	data, err := json.MarshalIndent(defs, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToOpenAIFormat 转换为 OpenAI Function Calling 格式
func (r *Registry) ToOpenAIFormat() []map[string]interface{} {
	defs := r.List()
	result := make([]map[string]interface{}, 0, len(defs))

	for _, def := range defs {
		properties := make(map[string]interface{})
		required := make([]string, 0)

		for _, param := range def.Parameters {
			properties[param.Name] = map[string]interface{}{
				"type":        param.Type,
				"description": param.Description,
			}
			if param.Required {
				required = append(required, param.Name)
			}
		}

		result = append(result, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        def.Name,
				"description": def.Description,
				"parameters": map[string]interface{}{
					"type":       "object",
					"properties": properties,
					"required":   required,
				},
			},
		})
	}

	return result
}
