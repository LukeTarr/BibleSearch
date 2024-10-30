package services

import (
	"BibleSearch/data"
	"BibleSearch/model"
	"BibleSearch/templates"
	"crypto/tls"
	"net/http"
	"strconv"
	"time"

	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ChromaService struct {
	Client        *chroma.Client
	ConfigService *ConfigService
	Collection    *chroma.Collection
}

func NewDefaultChromaService(configService *ConfigService) *ChromaService {
	client := chroma.NewClient(configService.ChromaURL)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client.ApiClient.GetConfig().HTTPClient = &http.Client{Transport: tr}
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

	embeddingFunction := openai.NewOpenAIEmbeddingFunction(c.ConfigService.OpenAIKey)

	for _, book := range *bookSlice {
		for chapterCounter, chapter := range book.Chapters {
			for verseCounter, verse := range chapter {
				metadatas := []map[string]interface{}{{
					"book":    book.Name,
					"chapter": strconv.Itoa(chapterCounter + 1),
					"verse":   strconv.Itoa(verseCounter + 1),
				}}

				embedding, err := embeddingFunction.CreateEmbedding([]string{verse})
				if err != nil {
					return err
				}

				successful := false
				for !successful {
					_, err := c.Collection.Add(embedding, metadatas, []string{verse}, []string{strconv.Itoa(globalCounter)})
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

func (c *ChromaService) getQueryResults(query string) (*[]model.ChromaQueryResultsDTO, error) {

	input := []string{query}
	n := int32(10)

	qr, err := c.query(input, n, nil, nil, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error querying")
		return nil, err
	}

	documents := qr.Documents[0]
	metaDatas := qr.Metadatas[0]
	ids := qr.Ids[0]
	distances := qr.Distances[0]

	resultSlice := make([]model.ChromaQueryResultsDTO, 0)

	for idx, doc := range documents {
		text := doc
		id := ids[idx]
		distance := float64(distances[idx])

		// The unquoting will never throw this error because we know the data in Chroma, so we can ignore the errors
		metaData := metaDatas[idx]
		book, _ := strconv.Unquote(string(metaData["book"].([]byte)))
		chapter, _ := strconv.Unquote(string(metaData["chapter"].([]byte)))
		verse, _ := strconv.Unquote(string(metaData["verse"].([]byte)))

		resultSlice = append(resultSlice, model.ChromaQueryResultsDTO{
			Metadata: model.Metadata{
				Book:          book,
				Chapter:       chapter,
				Verse:         verse,
				ReferenceLink: "https://www.bible.com/bible/1/" + data.BookAbbrevMap[book] + "." + chapter,
			},
			Distance: distance,
			Text:     text,
			Id:       id,
		})
	}

	return &resultSlice, nil
}

// HandleQueryRequest godoc
// @Summary query the vector database
// @Schemes
// @Description query the vector database
// @Tags query
// @Accept json
// @Produce json
// @Param query body model.QueryDTO true "query"
// @Success 200 {object} model.QueryResultsDTO
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

	log.Info().Str("query", queryDTO.Query).Msg("Received API query request")

	resultSlice, err := c.getQueryResults(queryDTO.Query)
	if err != nil {
		log.Error().Err(err).Msg("Error getting query results")
		ctx.JSON(500, model.ErrorDTO{
			Error: "error getting query results",
		})
		return
	}

	result := model.QueryResultsDTO{
		Result: *resultSlice,
	}

	ctx.JSON(200, result)
}

func (c *ChromaService) HandleHTMXQuery(ctx *gin.Context) {

	query := ctx.PostForm("query")
	log.Info().Str("query", query).Msg("Received HTMX query request")
	resultSlice, _ := c.getQueryResults(query)

	comp := templates.SearchResults(*resultSlice)
	ctx.Writer.Header().Set("Content-Type", "text/html")
	comp.Render(ctx.Request.Context(), ctx.Writer)
}
