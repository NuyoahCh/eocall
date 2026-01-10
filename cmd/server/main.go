package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NuyoahCh/eocall/internal/agent"
	"github.com/NuyoahCh/eocall/internal/agent/planner"
	"github.com/NuyoahCh/eocall/internal/chat/session"
	"github.com/NuyoahCh/eocall/internal/llm/eino"
	"github.com/NuyoahCh/eocall/internal/tools/registry"
	"github.com/NuyoahCh/eocall/pkg/config"
	"github.com/NuyoahCh/eocall/pkg/logger"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 初始化日志
	logger.Init(cfg.Log.Level, cfg.Log.Format)
	logger.Info("starting eocall server", "host", cfg.Server.Host, "port", cfg.Server.Port)

	// 初始化组件
	sessionMgr := session.NewManager(cfg.Session.TTL)
	toolRegistry := registry.NewRegistry()

	// 初始化 LLM 客户端
	llmClient, err := eino.NewClient(&eino.Config{
		Provider:    cfg.LLM.Provider,
		Model:       cfg.LLM.Model,
		APIKey:      cfg.LLM.APIKey,
		BaseURL:     cfg.LLM.BaseURL,
		MaxTokens:   cfg.LLM.MaxTokens,
		Temperature: cfg.LLM.Temperature,
	})
	if err != nil {
		log.Fatalf("failed to create llm client: %v", err)
	}

	// 初始化 Planner
	p := planner.NewPlanner(llmClient)

	// 初始化 Agent
	agentInstance := agent.NewAgent(p, toolRegistry, nil, sessionMgr, &agent.Config{
		MaxHistory:   cfg.Session.MaxHistory,
		SummaryAfter: cfg.Session.SummaryAfter,
	})

	// 设置 HTTP 路由
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/chat", chatHandler(agentInstance))
	mux.HandleFunc("/api/v1/chat/stream", streamChatHandler(agentInstance))

	// 启动服务器
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	// 优雅关闭
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down server...")
		server.Close()
	}()

	logger.Info("server started", "addr", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func chatHandler(a *agent.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: 实现聊天 API
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"chat endpoint"}`))
	}
}

func streamChatHandler(a *agent.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: 实现流式聊天 API (SSE)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Write([]byte("data: {\"message\":\"stream endpoint\"}\n\n"))
	}
}
