package config

import (
	"os"
	"sync"
)

type Config struct {
	AppID     string
	AppSecret string
}

var (
	instance *Config
	once     sync.Once
)

// GetConfig returns a singleton instance of Config with values from environment variables
func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{
			AppID:     getEnv("WECHAT_APPID", ""),
			AppSecret: getEnv("WECHAT_SECRET", ""),
		}
	})
	return instance
}

// getEnv reads an environment variable or returns a default value if not set
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
