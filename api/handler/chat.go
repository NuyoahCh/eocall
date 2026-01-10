package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/NuyoahCh/eocall/internal/agent"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	agent *agent.Agent
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(a *agent.Agent) *ChatHandler {
	return &ChatHandler{agent: a}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Handle 处理聊天请求
func (h *ChatHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == "" || req.Message == "" {
		writeError(w, http.StatusBadRequest, "user_id and message are required")
		return
	}

	if req.SessionID == "" {
		req.SessionID = "default"
	}

	resp, err := h.agent.Chat(r.Context(), &agent.Request{
		UserID:    req.UserID,
		SessionID: req.SessionID,
		Message:   req.Message,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ChatResponse{
		Code:    0,
		Message: "success",
		Data:    resp,
	})
}

// HandleStream 处理流式聊天请求 (SSE)
func (h *ChatHandler) HandleStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.UserID == "" || req.Message == "" {
		writeError(w, http.StatusBadRequest, "user_id and message are required")
		return
	}

	if req.SessionID == "" {
		req.SessionID = "default"
	}

	// 设置 SSE 头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	// 流式回调
	callback := func(chunk string) {
		data := map[string]string{"content": chunk}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		flusher.Flush()
	}

	err := h.agent.ChatStream(r.Context(), &agent.Request{
		UserID:    req.UserID,
		SessionID: req.SessionID,
		Message:   req.Message,
	}, callback)

	if err != nil {
		fmt.Fprintf(w, "data: {\"error\":\"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}

	// 发送结束标记
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ChatResponse{
		Code:    status,
		Message: message,
	})
}
