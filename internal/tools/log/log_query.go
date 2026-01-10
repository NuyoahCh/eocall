package log

import (
	"context"
	"fmt"
	"time"

	"github.com/NuyoahCh/eocall/internal/tools/registry"
)

// LogQueryTool 日志查询工具
type LogQueryTool struct {
	client LogClient
}

// LogClient 日志客户端接口
type LogClient interface {
	Query(ctx context.Context, req *LogQueryRequest) (*LogQueryResponse, error)
}

// LogQueryRequest 日志查询请求
type LogQueryRequest struct {
	Service   string    `json:"service"`
	Level     string    `json:"level"`
	Keyword   string    `json:"keyword"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Limit     int       `json:"limit"`
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Service   string    `json:"service"`
	Message   string    `json:"message"`
	TraceID   string    `json:"trace_id,omitempty"`
}

// LogQueryResponse 日志查询响应
type LogQueryResponse struct {
	Logs  []LogEntry `json:"logs"`
	Total int        `json:"total"`
}

// NewLogQueryTool 创建日志查询工具
func NewLogQueryTool(client LogClient) *LogQueryTool {
	return &LogQueryTool{client: client}
}

// Definition 工具定义
func (t *LogQueryTool) Definition() registry.ToolDefinition {
	return registry.ToolDefinition{
		Name:        "log_query",
		Description: "查询服务日志，支持按服务名、日志级别、关键词和时间范围筛选",
		Parameters: []registry.ToolParameter{
			{Name: "service", Type: "string", Description: "服务名称", Required: true},
			{Name: "level", Type: "string", Description: "日志级别: debug/info/warn/error", Required: false},
			{Name: "keyword", Type: "string", Description: "搜索关键词", Required: false},
			{Name: "start_time", Type: "string", Description: "开始时间 (RFC3339格式)", Required: false},
			{Name: "end_time", Type: "string", Description: "结束时间 (RFC3339格式)", Required: false},
			{Name: "limit", Type: "integer", Description: "返回条数限制，默认100", Required: false},
		},
	}
}

// Execute 执行日志查询
func (t *LogQueryTool) Execute(ctx context.Context, params map[string]interface{}) (*registry.ToolResult, error) {
	req := &LogQueryRequest{
		Limit: 100,
	}

	// 解析参数
	if v, ok := params["service"].(string); ok {
		req.Service = v
	} else {
		return &registry.ToolResult{
			Success: false,
			Error:   "service is required",
		}, nil
	}

	if v, ok := params["level"].(string); ok {
		req.Level = v
	}
	if v, ok := params["keyword"].(string); ok {
		req.Keyword = v
	}
	if v, ok := params["limit"].(float64); ok {
		req.Limit = int(v)
	}
	if v, ok := params["start_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			req.StartTime = t
		}
	}
	if v, ok := params["end_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			req.EndTime = t
		}
	}

	// 设置默认时间范围
	if req.EndTime.IsZero() {
		req.EndTime = time.Now()
	}
	if req.StartTime.IsZero() {
		req.StartTime = req.EndTime.Add(-1 * time.Hour)
	}

	// 执行查询
	resp, err := t.client.Query(ctx, req)
	if err != nil {
		return &registry.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("query failed: %v", err),
		}, nil
	}

	return &registry.ToolResult{
		Success: true,
		Data:    resp,
	}, nil
}
