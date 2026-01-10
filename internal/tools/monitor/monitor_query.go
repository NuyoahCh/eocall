package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/NuyoahCh/eocall/internal/tools/registry"
)

// MonitorQueryTool 监控查询工具
type MonitorQueryTool struct {
	client MonitorClient
}

// MonitorClient 监控客户端接口
type MonitorClient interface {
	QueryMetrics(ctx context.Context, req *MetricsQueryRequest) (*MetricsQueryResponse, error)
	QueryAlerts(ctx context.Context, req *AlertsQueryRequest) (*AlertsQueryResponse, error)
}

// MetricsQueryRequest 指标查询请求
type MetricsQueryRequest struct {
	Service   string    `json:"service"`
	Metric    string    `json:"metric"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Step      string    `json:"step"`
}

// MetricPoint 指标数据点
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// MetricsQueryResponse 指标查询响应
type MetricsQueryResponse struct {
	Metric string        `json:"metric"`
	Points []MetricPoint `json:"points"`
}

// AlertsQueryRequest 告警查询请求
type AlertsQueryRequest struct {
	Service  string `json:"service"`
	Severity string `json:"severity"`
	Status   string `json:"status"`
	Limit    int    `json:"limit"`
}

// Alert 告警
type Alert struct {
	ID          string    `json:"id"`
	Service     string    `json:"service"`
	Name        string    `json:"name"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	StartTime   time.Time `json:"start_time"`
	ResolveTime time.Time `json:"resolve_time,omitempty"`
}

// AlertsQueryResponse 告警查询响应
type AlertsQueryResponse struct {
	Alerts []Alert `json:"alerts"`
	Total  int     `json:"total"`
}

// NewMonitorQueryTool 创建监控查询工具
func NewMonitorQueryTool(client MonitorClient) *MonitorQueryTool {
	return &MonitorQueryTool{client: client}
}

// Definition 工具定义
func (t *MonitorQueryTool) Definition() registry.ToolDefinition {
	return registry.ToolDefinition{
		Name:        "monitor_metrics",
		Description: "查询服务监控指标，如 CPU、内存、QPS、延迟等",
		Parameters: []registry.ToolParameter{
			{Name: "service", Type: "string", Description: "服务名称", Required: true},
			{Name: "metric", Type: "string", Description: "指标名称: cpu/memory/qps/latency/error_rate", Required: true},
			{Name: "start_time", Type: "string", Description: "开始时间 (RFC3339格式)", Required: false},
			{Name: "end_time", Type: "string", Description: "结束时间 (RFC3339格式)", Required: false},
			{Name: "step", Type: "string", Description: "采样间隔: 1m/5m/1h", Required: false},
		},
	}
}

// Execute 执行监控查询
func (t *MonitorQueryTool) Execute(ctx context.Context, params map[string]interface{}) (*registry.ToolResult, error) {
	req := &MetricsQueryRequest{
		Step: "1m",
	}

	if v, ok := params["service"].(string); ok {
		req.Service = v
	} else {
		return &registry.ToolResult{Success: false, Error: "service is required"}, nil
	}

	if v, ok := params["metric"].(string); ok {
		req.Metric = v
	} else {
		return &registry.ToolResult{Success: false, Error: "metric is required"}, nil
	}

	if v, ok := params["step"].(string); ok {
		req.Step = v
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

	if req.EndTime.IsZero() {
		req.EndTime = time.Now()
	}
	if req.StartTime.IsZero() {
		req.StartTime = req.EndTime.Add(-1 * time.Hour)
	}

	resp, err := t.client.QueryMetrics(ctx, req)
	if err != nil {
		return &registry.ToolResult{Success: false, Error: fmt.Sprintf("query failed: %v", err)}, nil
	}

	return &registry.ToolResult{Success: true, Data: resp}, nil
}

// AlertQueryTool 告警查询工具
type AlertQueryTool struct {
	client MonitorClient
}

// NewAlertQueryTool 创建告警查询工具
func NewAlertQueryTool(client MonitorClient) *AlertQueryTool {
	return &AlertQueryTool{client: client}
}

// Definition 工具定义
func (t *AlertQueryTool) Definition() registry.ToolDefinition {
	return registry.ToolDefinition{
		Name:        "alert_query",
		Description: "查询服务告警信息",
		Parameters: []registry.ToolParameter{
			{Name: "service", Type: "string", Description: "服务名称", Required: false},
			{Name: "severity", Type: "string", Description: "告警级别: critical/warning/info", Required: false},
			{Name: "status", Type: "string", Description: "告警状态: firing/resolved", Required: false},
			{Name: "limit", Type: "integer", Description: "返回条数限制", Required: false},
		},
	}
}

// Execute 执行告警查询
func (t *AlertQueryTool) Execute(ctx context.Context, params map[string]interface{}) (*registry.ToolResult, error) {
	req := &AlertsQueryRequest{Limit: 50}

	if v, ok := params["service"].(string); ok {
		req.Service = v
	}
	if v, ok := params["severity"].(string); ok {
		req.Severity = v
	}
	if v, ok := params["status"].(string); ok {
		req.Status = v
	}
	if v, ok := params["limit"].(float64); ok {
		req.Limit = int(v)
	}

	resp, err := t.client.QueryAlerts(ctx, req)
	if err != nil {
		return &registry.ToolResult{Success: false, Error: fmt.Sprintf("query failed: %v", err)}, nil
	}

	return &registry.ToolResult{Success: true, Data: resp}, nil
}
