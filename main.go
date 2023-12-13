package main

import (
	"BibleSearch/config"
	"BibleSearch/data"
	"BibleSearch/services"
	"fmt"
	"github.com/rs/zerolog/log"
)

func main() {
	config.ReadDotEnv()
	configuration := config.NewDefaultConfig()

	log.Info().Msg("Creating Chroma Service and Collection")
	chromaService := services.NewDefaultChromaService(configuration)
	collection, err := chromaService.CreateCollection("bible")
	if err != nil {
		log.Error().Err(err).Msg("Error creating collection")
		return
	}

	bookSlice, err := data.GetBookSlice()
	if err != nil {
		log.Error().Err(err).Msg("Error getting book slice")
		return
	}

	log.Info().Msg("Adding Documents to Collection")
	err = chromaService.AddBooksToCollection(bookSlice)
	if err != nil {
		log.Error().Err(err).Msg("Error adding books to collection")
		return
	}

	countDocs, err := chromaService.Collection.Count()
	if err != nil {
		log.Error().Err(err).Msg("Error `querying documents")
		return
	}

	log.Info().Int32("docsCounter", countDocs).Msg("Counted documents")

	qr, err := collection.Query([]string{"Formless, void, emptiness"}, 1, nil, nil, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error `querying documents")
		return
	}

	fmt.Println(string(qr.Metadatas[0][0]["book"].([]byte)))
	fmt.Println(string(qr.Metadatas[0][0]["chapter"].([]byte)))
	fmt.Println(string(qr.Metadatas[0][0]["verse"].([]byte)))
}
