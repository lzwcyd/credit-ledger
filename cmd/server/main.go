package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/credit-ledger/config"
	"github.com/yourorg/credit-ledger/internal/api"
	"github.com/yourorg/credit-ledger/pkg/database"
	"github.com/yourorg/credit-ledger/pkg/cache"
	"github.com/yourorg/credit-ledger/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// 初始化日志
	logger.Init(cfg.Log.Level, cfg.Log.Format)
	log := logger.GetLogger()

	// 初始化数据库（开发环境使用 SQLite）
	db, err := database.NewMySQL(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// 初始化缓存（可选）
	redisClient, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		log.Warn("Failed to connect to Redis, continuing without cache", zap.Error(err))
		redisClient = nil
	} else {
		defer redisClient.Close()
	}

	// 初始化 API 路由
	router := api.NewRouter(db, redisClient)

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// 启动服务器
	go func() {
		log.Info("Server starting", zap.Int("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}