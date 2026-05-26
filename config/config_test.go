package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "3307")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("REDIS_HOST", "redishost")
	os.Setenv("REDIS_PORT", "6380")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "console")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 验证服务器配置
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected server port 9090, got %d", cfg.Server.Port)
	}

	// 验证数据库配置
	if cfg.Database.Host != "testhost" {
		t.Errorf("Expected DB host 'testhost', got '%s'", cfg.Database.Host)
	}
	if cfg.Database.Port != 3307 {
		t.Errorf("Expected DB port 3307, got %d", cfg.Database.Port)
	}
	if cfg.Database.User != "testuser" {
		t.Errorf("Expected DB user 'testuser', got '%s'", cfg.Database.User)
	}

	// 验证 Redis 配置
	if cfg.Redis.Host != "redishost" {
		t.Errorf("Expected Redis host 'redishost', got '%s'", cfg.Redis.Host)
	}
	if cfg.Redis.Port != 6380 {
		t.Errorf("Expected Redis port 6380, got %d", cfg.Redis.Port)
	}

	// 验证日志配置
	if cfg.Log.Level != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", cfg.Log.Level)
	}
	if cfg.Log.Format != "console" {
		t.Errorf("Expected log format 'console', got '%s'", cfg.Log.Format)
	}
}