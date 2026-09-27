package services

import (
	"BibleSearch/data"
	"BibleSearch/model"
	"BibleSearch/templates"
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Changing the embedding model requires a new collection, since vectors from different models aren't comparable
const (
	EmbeddingModel = "text-embedding-3-small"
	CollectionName = "bible-" + EmbeddingModel
)

// Verses per OpenAI embeddings request and Chroma add. OpenAI accepts up to 2048 inputs per request.
const vectorizeBatchSize = 1000

// Batches embedded and added at the same time. Higher values may run into OpenAI's tokens-per-minute limit,
// which the retry loop absorbs.
const vectorizeConcurrency = 4

type ChromaService struct {
	ConfigService *ConfigService
	CollectionID  string
	chroma        *chromaClient
	openai        *OpenAIClient
}

func NewDefaultChromaService(configService *ConfigService) *ChromaService {
	return &ChromaService{
		ConfigService: configService,
		chroma:        newChromaClient(configService.ChromaURL),
		openai:        NewOpenAIClient(configService.OpenAIKey, EmbeddingModel),
	}
}

func (c *ChromaService) ResetClient() error {
	return c.chroma.reset(context.Background())
}

func (c *ChromaService) CreateCollection(collectionName string) error {
	id, err := c.chroma.getOrCreateCollection(context.Background(), collectionName)
	if err != nil {
		return err
	}

	c.CollectionID = id
	return nil
}

func (c *ChromaService) Count() (int, error) {
	return c.chroma.count(context.Background(), c.CollectionID)
}

func (c *ChromaService) AddBooksToCollection(bookSlice *[]model.Book) error {
	var ids, documents []string
	var metadatas []verseMetadata

	for _, book := range *bookSlice {
		for chapterCounter, chapter := range book.Chapters {
			for verseCounter, verse := range chapter {
				ids = append(ids, strconv.Itoa(len(ids)))
				documents = append(documents, verse)
				metadatas = append(metadatas, verseMetadata{
					Book:    book.Name,
					Chapter: strconv.Itoa(chapterCounter + 1),
					Verse:   strconv.Itoa(verseCounter + 1),
				})
			}
		}
	}

	ctx := context.Background()
	// Buffered channel as a counting semaphore: at most vectorizeConcurrency batches in flight
	sem := make(chan struct{}, vectorizeConcurrency)
	var wg sync.WaitGroup
	var added atomic.Int64

	for start := 0; start < len(ids); start += vectorizeBatchSize {
		end := min(start+vectorizeBatchSize, len(ids))

		sem <- struct{}{}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			defer func() { <-sem }()

			for {
				err := c.addBatch(ctx, ids[start:end], documents[start:end], metadatas[start:end])
				if err == nil {
					break
				}
				log.Error().Err(err).Int("start", start).Msg("Error adding documents, retrying")
				time.Sleep(5 * time.Second)
			}
			log.Info().Int64("added", added.Add(int64(end-start))).Int("total", len(ids)).Msg("Added batch")
		}(start, end)
	}

	wg.Wait()
	return nil
}

func (c *ChromaService) addBatch(ctx context.Context, ids []string, documents []string, metadatas []verseMetadata) error {
	embeddings, err := c.openai.Embed(ctx, documents)
	if err != nil {
		return err
	}
	return c.chroma.add(ctx, c.CollectionID, ids, embeddings, documents, metadatas)
}

func (c *ChromaService) getQueryResults(ctx context.Context, query string) (*[]model.ChromaQueryResultsDTO, error) {
	embeddings, err := c.openai.Embed(ctx, []string{query})
	if err != nil {
		log.Error().Err(err).Msg("Error embedding query")
		return nil, err
	}

	qr, err := c.chroma.query(ctx, c.CollectionID, embeddings[0], 10)
	if err != nil {
		log.Error().Err(err).Msg("Error querying")
		return nil, err
	}

	resultSlice := make([]model.ChromaQueryResultsDTO, 0)
	if len(qr.IDs) == 0 {
		return &resultSlice, nil
	}

	documents := qr.Documents[0]
	metaDatas := qr.Metadatas[0]
	ids := qr.IDs[0]
	distances := qr.Distances[0]

	for idx, doc := range documents {
		metaData := metaDatas[idx]
		resultSlice = append(resultSlice, model.ChromaQueryResultsDTO{
			Metadata: model.Metadata{
				Book:          metaData.Book,
				Chapter:       metaData.Chapter,
				Verse:         metaData.Verse,
				ReferenceLink: "https://www.bible.com/bible/1/" + data.BookAbbrevMap[metaData.Book] + "." + metaData.Chapter,
			},
			Distance: distances[idx],
			Text:     doc,
			Id:       ids[idx],
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

	resultSlice, err := c.getQueryResults(ctx.Request.Context(), queryDTO.Query)
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
	resultSlice, err := c.getQueryResults(ctx.Request.Context(), query)
	if err != nil {
		log.Error().Err(err).Msg("Error getting query results")
		ctx.String(500, "error getting query results")
		return
	}

	comp := templates.SearchResults(*resultSlice)
	ctx.Writer.Header().Set("Content-Type", "text/html")
	comp.Render(ctx.Request.Context(), ctx.Writer)
}
