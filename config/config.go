package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config はアプリケーションの設定を管理する構造体
type Config struct {
	// RSS フィード関連
	FeedURL         string
	CheckInterval   time.Duration
	
	// DeepL API 関連
	DeepLAPIKey     string
	DeepLAPIURL     string
	
	// OpenAI API 関連
	OpenAIAPIKey    string
	OpenAIModel     string
	
	// Slack 関連
	SlackWebhookURL string
	SlackChannel    string
	SlackUseThreads bool
	
	// アプリケーション設定
	LogLevel        string
	Timezone        string
}

// LoadConfig は環境変数から設定を読み込む
func LoadConfig() *Config {
	// .envファイルを読み込み（存在する場合）
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	config := &Config{
		// RSS フィード関連
		FeedURL:         getEnvOrDefault("FEED_URL", "https://blog.bytebytego.com/feed"),
		CheckInterval:   getDurationFromEnv("CHECK_INTERVAL_MINUTES", 30) * time.Minute,
		
		// DeepL API 関連
		DeepLAPIKey:     getEnvOrPanic("DEEPL_API_KEY"),
		DeepLAPIURL:     getEnvOrDefault("DEEPL_API_URL", "https://api-free.deepl.com/v2/translate"),
		
		// OpenAI API 関連
		OpenAIAPIKey:    getEnvOrPanic("OPENAI_API_KEY"),
		OpenAIModel:     getEnvOrDefault("OPENAI_MODEL", "gpt-3.5-turbo"),
		
		// Slack 関連
		SlackWebhookURL: getEnvOrPanic("SLACK_WEBHOOK_URL"),
		SlackChannel:    getEnvOrDefault("SLACK_CHANNEL", "#general"),
		SlackUseThreads: getBoolFromEnv("SLACK_USE_THREADS", true),
		
		// アプリケーション設定
		LogLevel:        getEnvOrDefault("LOG_LEVEL", "info"),
		Timezone:        getEnvOrDefault("TIMEZONE", "Asia/Tokyo"),
	}

	// 設定値の検証
	if err := config.validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	return config
}

// validate は設定値の妥当性をチェックする
func (c *Config) validate() error {
	if c.FeedURL == "" {
		return fmt.Errorf("FEED_URL is required")
	}
	if c.DeepLAPIKey == "" {
		return fmt.Errorf("DEEPL_API_KEY is required")
	}
	if c.OpenAIAPIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}
	if c.SlackWebhookURL == "" {
		return fmt.Errorf("SLACK_WEBHOOK_URL is required")
	}
	if c.CheckInterval <= 0 {
		return fmt.Errorf("CHECK_INTERVAL_MINUTES must be greater than 0")
	}
	return nil
}

// getEnvOrDefault は環境変数の値を取得し、存在しない場合はデフォルト値を返す
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvOrPanic は環境変数の値を取得し、存在しない場合はパニックを起こす
func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return value
}

// getDurationFromEnv は環境変数から数値を取得し、time.Durationに変換する
func getDurationFromEnv(key string, defaultValue int) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return time.Duration(defaultValue)
	}
	
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid value for %s, using default: %d", key, defaultValue)
		return time.Duration(defaultValue)
	}
	
	return time.Duration(value)
}

// getBoolFromEnv は環境変数からブール値を取得する
func getBoolFromEnv(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	
	// 各種真値パターンをサポート
	switch strings.ToLower(valueStr) {
	case "true", "t", "yes", "y", "1", "on", "enable", "enabled":
		return true
	case "false", "f", "no", "n", "0", "off", "disable", "disabled":
		return false
	default:
		log.Printf("Warning: Invalid boolean value for %s: %s, using default: %t", key, valueStr, defaultValue)
		return defaultValue
	}
}