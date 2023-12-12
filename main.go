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

	qr, err := collection.Query([]string{"The first human"}, 1, nil, nil, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error `querying documents")
		return
	}

	log.Info().Any("qr", qr.Documents).Msg("Query Results")
	log.Info().Any("qr", qr.Ids).Msg("Query Results")
	log.Info().Any("qr", qr.Distances).Msg("Query Results")
	log.Info().Any("qr", qr.Metadatas).Msg("Query Results")
	fmt.Println(qr.Metadatas[0][0])

}
