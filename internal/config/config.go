package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken string
	DatabasePath string
	DeepSeekAPIKey string
	DailyTokenLimit int
	DailyPromptLimit int
	TokenPrice float64
}

func Load() (*Config, error){
	_ = godotenv.Load(".env")

	return &Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		DatabasePath: os.Getenv("DATABASE_PATH"),
		DeepSeekAPIKey: os.Getenv("DEEPSEEK_API_KEY"),
		DailyTokenLimit: 60000,
		DailyPromptLimit: 80,
		TokenPrice: 0.001,
	}, nil
}
