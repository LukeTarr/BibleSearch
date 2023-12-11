package config

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"os"
)

type Config struct {
	ChromaURL string
	OpenAIKey string
}

func NewDefaultConfig() *Config {
	return &Config{
		ChromaURL: os.Getenv("CHROMA_URL"),
		OpenAIKey: os.Getenv("OPENAI_API_KEY"),
	}
}

func ReadDotEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Error().Err(err).Msg("Error loading .env file")
	}
}
