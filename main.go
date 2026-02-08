package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// 配置结构体
type Config struct {
	LogInterval int    // 日志间隔(秒)
	ServerPort  int    // HTTP服务端口
	LogMessage  string // 日志消息
	IncludeHTTP bool   // 是否包含HTTP服务
	Hostname    string // 主机名
	StartTime   time.Time
}

var config Config

func init() {
	// 从环境变量读取配置，无则使用默认值
	interval, _ := strconv.Atoi(getEnv("LOG_INTERVAL", "1"))
	port, _ := strconv.Atoi(getEnv("SERVER_PORT", "8080"))

	config = Config{
		LogInterval: interval,
		ServerPort:  port,
		LogMessage:  getEnv("LOG_MESSAGE", "Logger is running"),
		IncludeHTTP: getEnv("INCLUDE_HTTP", "true") == "true",
		Hostname:    getEnv("HOSTNAME", "unknown"),
		StartTime:   time.Now(),
	}

	log.Printf("Logger配置加载完成: 间隔=%d秒, 端口=%d, HTTP服务=%v",
		config.LogInterval, config.ServerPort, config.IncludeHTTP)
	log.Printf("主机名: %s", config.Hostname)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 健康检查端点
func healthHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(config.StartTime)
	_ = map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"hostname":  config.Hostname,
		"uptime":    uptime.String(),
		"config": map[string]interface{}{
			"log_interval": config.LogInterval,
			"log_message":  config.LogMessage,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s","hostname":"%s","uptime":"%s"}`,
		time.Now().Format(time.RFC3339), config.Hostname, uptime.String())
}

// 主日志循环
func startLogging() {
	ticker := time.NewTicker(time.Duration(config.LogInterval) * time.Second)
	defer ticker.Stop()

	logCount := 0
	for range ticker.C {
		logCount++
		uptime := time.Since(config.StartTime)
		log.Printf("[%s] %s | 计数: %d | 运行: %s | 主机: %s",
			time.Now().Format("2006-01-02 15:04:05"),
			config.LogMessage,
			logCount,
			uptime.Truncate(time.Second).String(),
			config.Hostname)

		// 模拟不同级别的日志
		if logCount%10 == 0 {
			log.Printf("INFO: 已记录 %d 条日志消息", logCount)
		}
		if logCount%50 == 0 {
			log.Printf("WARN: 这是一个警告级别的日志示例")
		}
	}
}

// 启动HTTP服务器
func startHTTPServer() {
	if !config.IncludeHTTP {
		log.Println("HTTP服务已禁用")
		return
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		uptime := time.Since(config.StartTime)
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head><title>Go Logger</title></head>
<body>
	<h1>Go Logger Demo</h1>
	<p>状态: <span style="color:green">✅ 运行中</span></p>
	<p>启动时间: %s</p>
	<p>运行时长: %s</p>
	<p>主机名: %s</p>
	<p>日志间隔: %d秒</p>
	<p>总日志数: 持续增加中...</p>
	<p><a href="/health">健康检查端点</a></p>
</body>
</html>`,
			config.StartTime.Format("2006-01-02 15:04:05"),
			uptime.Truncate(time.Second).String(),
			config.Hostname,
			config.LogInterval)
	})

	addr := fmt.Sprintf(":%d", config.ServerPort)
	log.Printf("HTTP服务器启动，监听端口 %d", config.ServerPort)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("HTTP服务器启动失败: %v", err)
	}
}

func main() {
	log.Println("🚀 Go Logger Demo 启动中...")
	log.Println("📋 配置信息:")
	log.Printf("  - 日志间隔: %d秒", config.LogInterval)
	log.Printf("  - 日志消息: %s", config.LogMessage)
	log.Printf("  - HTTP端口: %d", config.ServerPort)
	log.Printf("  - 包含HTTP服务: %v", config.IncludeHTTP)
	log.Printf("  - 主机名: %s", config.Hostname)
	log.Println("🔧 按 Ctrl+C 停止程序")

	// 启动HTTP服务（如果启用）
	if config.IncludeHTTP {
		go startHTTPServer()
	}

	fmt.Println("Hello, World!")
	// 开始日志循环
	startLogging()
}
