package main

import (
	"BibleSearch/config"
	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	config.ReadDotEnv()
	configuration := config.NewDefaultConfig()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	log.Info().Msg("Creating Chroma Client")
	client := chroma.NewClient(configuration.ChromaURL)
	meta := map[string]interface{}{}

	log.Info().Msg("Creating OpenAI Embedding Function")
	embeddingFunction := openai.NewOpenAIEmbeddingFunction(configuration.OpenAIKey) //create a new OpenAI Embedding function

	log.Info().Msg("Creating Collection")
	collection, err := client.CreateCollection("test", meta, true, embeddingFunction, chroma.L2)
	if err != nil {
		log.Error().Err(err).Msg("Error creating collection")
		return
	}

	log.Info().Msg("Adding Documents to Collection")
	documents := []string{
		"This is a document about cats. Cats are great.",
		"this is a document about dogs. Dogs are great.",
	}
	ids := []string{
		"ID1",
		"ID2",
	}

	metadatas := []map[string]interface{}{
		{"key1": "value1"},
		{"key2": "value2"},
	}

	log.Info().Msg("Adding Documents to Collection")
	col, addError := collection.Add(nil, metadatas, documents, ids)
	if addError != nil {
		log.Error().Err(addError).Msg("Error adding documents")
		return
	}

	log.Info().Any("col", col).Msg("Added Documents to Collection")

	countDocs, qrerr := collection.Count()
	if qrerr != nil {
		log.Error().Err(qrerr).Msg("Error `querying documents")
		return
	}

	log.Info().Int32("docsCounter", countDocs).Msg("Counted documents")

	qr, qrerr := collection.Query([]string{"I love dogs"}, 5, nil, nil, nil)
	if qrerr != nil {
		log.Error().Err(qrerr).Msg("Error `querying documents")
		return
	}

	log.Info().Any("qr", qr).Msg("Query Results")
}
