package services

import (
	"BibleSearch/model"
	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

type ChromaService struct {
	Client        *chroma.Client
	ConfigService *ConfigService
	Collection    *chroma.Collection
}

func NewDefaultChromaService(configService *ConfigService) *ChromaService {
	client := chroma.NewClient(configService.ChromaURL)
	return &ChromaService{
		Client:        client,
		ConfigService: configService,
		Collection:    nil,
	}
}

func (c *ChromaService) ResetClient() error {
	_, err := c.Client.Reset()
	if err != nil {
		return err
	}
	return nil
}

func (c *ChromaService) CreateCollection(collectionName string) (*chroma.Collection, error) {
	meta := map[string]interface{}{}
	embeddingFunction := openai.NewOpenAIEmbeddingFunction(c.ConfigService.OpenAIKey)
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
		for chapterCounter, chapter := range book.Chapters {
			for verseCounter, verse := range chapter {
				metadatas := []map[string]interface{}{{
					"book":    book.Name,
					"chapter": strconv.Itoa(chapterCounter + 1),
					"verse":   strconv.Itoa(verseCounter + 1),
				}}

				successful := false
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
			}
		}
	}
	return nil
}

func (c *ChromaService) query(text []string, n int32, where map[string]interface{}, whereDocuments map[string]interface{}, include []chroma.QueryEnum) (*chroma.QueryResults, error) {
	qr, err := c.Collection.Query(text, n, where, whereDocuments, include)
	if err != nil {
		return nil, err
	}
	return qr, nil
}

// HandleQueryRequest godoc
// @Summary query the vector database
// @Schemes
// @Description query the vector database
// @Tags query
// @Accept json
// @Produce json
// @Param query body model.QueryDTO true "query"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {string} model.ErrorDTO
// @Router /query [post]
func (c *ChromaService) HandleQueryRequest(ctx *gin.Context) {

	var queryDTO model.QueryDTO
	err := ctx.ShouldBindJSON(&queryDTO)
	if err != nil {
		log.Error().Err(err).Msg("Error binding json")
		ctx.JSON(500, model.ErrorDTO{
			Error: "error binding json",
		})
		return
	}

	input := []string{queryDTO.Query}
	n := int32(10)

	qr, err := c.query(input, n, nil, nil, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error querying collection")
		ctx.JSON(500, model.ErrorDTO{
			Error: "error querying collection",
		})
		return
	}
	//
	//documents := qr.Documents[0]
	//metaDatas := qr.Metadatas[0]
	//ids := qr.Ids[0]
	//
	//for idx, doc := range documents {
	//	text := doc
	//	metaData := metaDatas[idx]
	//
	//	id := ids[idx]
	//
	//
	//}

	ctx.JSON(200, qr)
}
