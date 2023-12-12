package services

import (
	"BibleSearch/config"
	"BibleSearch/model"
	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

type ChromaService struct {
	Client        *chroma.Client
	Configuration *config.Config
	Collection    *chroma.Collection
}

func NewDefaultChromaService(configuration *config.Config) *ChromaService {
	client := chroma.NewClient(configuration.ChromaURL)
	return &ChromaService{
		Client:        client,
		Configuration: configuration,
		Collection:    nil,
	}
}

func (c *ChromaService) CreateCollection(collectionName string) (*chroma.Collection, error) {
	meta := map[string]interface{}{}
	embeddingFunction := openai.NewOpenAIEmbeddingFunction(c.Configuration.OpenAIKey)
	collection, err := c.Client.CreateCollection(collectionName, meta, true, embeddingFunction, chroma.L2)
	if err != nil {
		return nil, err
	}

	c.Collection = collection
	return collection, nil
}

func (c *ChromaService) AddBooksToCollection(bookSlice *[]model.Book) error {
	globalCounter := 0
	for _, book := range *bookSlice {
		chapterCounter := 0
		for _, chapter := range book.Chapters {
			verseCounter := 0
			for _, verse := range chapter {

				md := map[string]interface{}{
					"book":    book.Name,
					"chapter": chapterCounter,
					"verse":   verseCounter,
				}
				metadatas := []map[string]interface{}{md}

				successful := false
				// keep retrying until it works

				for !successful {
					_, err := c.Collection.Add(nil, metadatas, []string{verse}, []string{strconv.Itoa(globalCounter)})
					if err != nil {
						log.Error().Err(err).Msg("Error adding documents, retrying")
						time.Sleep(5 * time.Second)
					} else {
						successful = true
					}
				}
				globalCounter++
				chapterCounter++
				verseCounter++
			}
		}
	}
	return nil
}

func (c *ChromaService) Query(text []string, n int32, where map[string]interface{}, whereDocuments map[string]interface{}, include []chroma.QueryEnum) (*chroma.QueryResults, error) {
	qr, err := c.Collection.Query(text, n, where, whereDocuments, include)
	if err != nil {
		log.Error().Err(err).Msg("Error `querying documents")
		return nil, err
	}
	return qr, nil
}
