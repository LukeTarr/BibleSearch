package services

import (
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
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
		log.Info().Msg("No .env file found, using environment variables")
	}
}
