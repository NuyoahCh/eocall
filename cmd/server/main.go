package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NuyoahCh/eocall/api/handler"
	"github.com/NuyoahCh/eocall/api/middleware"
	"github.com/NuyoahCh/eocall/internal/agent"
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

	ctx := context.Background()

	// 初始化 LLM 客户端 (使用 QuickChat 模型)
	llmClient, err := eino.NewClient(ctx, &eino.Config{
		APIKey:  cfg.LLM.QuickChat.APIKey,
		BaseURL: cfg.LLM.QuickChat.BaseURL,
		Model:   cfg.LLM.QuickChat.Model,
	})
	if err != nil {
		log.Fatalf("failed to create llm client: %v", err)
	}
	logger.Info("llm client initialized", "model", cfg.LLM.QuickChat.Model)

	// 初始化组件
	sessionMgr := session.NewManager(cfg.Session.TTL)
	toolRegistry := registry.NewRegistry()

	// 初始化 Agent
	agentInstance := agent.NewAgent(llmClient, toolRegistry, nil, sessionMgr, &agent.Config{
		MaxHistory:   cfg.Session.MaxHistory,
		SummaryAfter: cfg.Session.SummaryAfter,
	})

	// 创建处理器
	chatHandler := handler.NewChatHandler(agentInstance)

	// 设置 HTTP 路由
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/chat", chatHandler.Handle)
	mux.HandleFunc("/api/v1/chat/stream", chatHandler.HandleStream)

	// 应用中间件
	h := middleware.Chain(mux,
		middleware.Recovery,
		middleware.Logger,
		middleware.CORS,
	)

	// 启动服务器
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
