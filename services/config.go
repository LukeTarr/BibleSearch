package services

import (
	"github.com/joho/godotenv"
	"log/slog"
	"os"
)

type ConfigService struct {
	ChromaURL             string
	OpenAIKey             string
	VectorizationPassword string
	// Dev loads templates from disk on every request instead of from the binary
	Dev bool
}

func NewDefaultConfig() *ConfigService {
	return &ConfigService{
		ChromaURL:             os.Getenv("CHROMA_URL"),
		OpenAIKey:             os.Getenv("OPENAI_API_KEY"),
		VectorizationPassword: os.Getenv("VECTORIZATION_PASSWORD"),
		Dev:                   os.Getenv("DEV") == "true",
	}
}

func ReadDotEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		slog.Info("No .env file found, using environment variables")
	}
}
