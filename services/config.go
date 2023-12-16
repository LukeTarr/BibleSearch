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
}

func NewDefaultConfig() *ConfigService {
	return &ConfigService{
		ChromaURL:             os.Getenv("CHROMA_URL"),
		OpenAIKey:             os.Getenv("OPENAI_API_KEY"),
		VectorizationPassword: os.Getenv("VECTORIZATION_PASSWORD"),
	}
}

func ReadDotEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Error().Err(err).Msg("Error loading .env file")
	}
}
