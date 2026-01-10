package registry

import (
	"context"
	"testing"
)

// MockTool 测试用工具
type MockTool struct {
	name string
}

func (t *MockTool) Definition() ToolDefinition {
	return ToolDefinition{
		Name:        t.name,
		Description: "mock tool for testing",
		Parameters: []ToolParameter{
			{Name: "param1", Type: "string", Description: "test param", Required: true},
		},
	}
}

func (t *MockTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	return &ToolResult{
		Success: true,
		Data:    map[string]string{"result": "ok"},
	}, nil
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	tool := &MockTool{name: "test_tool"}
	r.Register(tool)

	got, err := r.Get("test_tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Definition().Name != "test_tool" {
		t.Errorf("expected test_tool, got %s", got.Definition().Name)
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()

	r.Register(&MockTool{name: "tool1"})
	r.Register(&MockTool{name: "tool2"})

	defs := r.List()
	if len(defs) != 2 {
		t.Errorf("expected 2 tools, got %d", len(defs))
	}
}

func TestRegistry_Execute(t *testing.T) {
	r := NewRegistry()
	r.Register(&MockTool{name: "test_tool"})

	result, err := r.Execute(context.Background(), "test_tool", map[string]interface{}{
		"param1": "value1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected success")
	}
}

func TestRegistry_ToOpenAIFormat(t *testing.T) {
	r := NewRegistry()
	r.Register(&MockTool{name: "test_tool"})

	format := r.ToOpenAIFormat()
	if len(format) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(format))
	}

	if format[0]["type"] != "function" {
		t.Error("expected type to be function")
	}
}
