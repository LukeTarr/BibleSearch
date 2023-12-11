package services

import (
	"BibleSearch/config"
	chroma "github.com/amikos-tech/chroma-go"
)

type ChromaService struct {
	client        *chroma.Client
	configuration *config.Config
}

func NewDefaultChromaService(configuration *config.Config) *ChromaService {
	client := chroma.NewClient(configuration.ChromaURL)
	return &ChromaService{
		client:        client,
		configuration: configuration,
	}
}
