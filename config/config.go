package config

import (
	"os"
	"sync"
)

type Config struct {
	AppID        string
	AppSecret    string
	MongoURI     string
	MongoDB      string
	MongoUser    string
	MongoPass    string
	MongoTimeout int
}

var (
	instance *Config
	once     sync.Once
)

// GetConfig returns a singleton instance of Config with values from environment variables
func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{
			AppID:        getEnv("WECHAT_APPID", ""),
			AppSecret:    getEnv("WECHAT_SECRET", ""),
			MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
			MongoDB:      getEnv("MONGO_DB", "playtime"),
			MongoUser:    getEnv("MONGO_USER", "admin"),
			MongoPass:    getEnv("MONGO_PASS", "admin"),
			MongoTimeout: 10, // 10 seconds timeout
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
