package services

import (
	"BibleSearch/data"
	"BibleSearch/model"
	"BibleSearch/templates"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// Changing the embedding model requires a new collection, since vectors from different models aren't comparable
const (
	EmbeddingModel = "text-embedding-3-small"
	CollectionName = "bible-" + EmbeddingModel
)

// Verses per OpenAI embeddings request and Chroma add. OpenAI accepts up to 2048 inputs per request.
const vectorizeBatchSize = 1000

// Longest search query accepted, in characters
const maxQueryLength = 500

var (
	errEmptyQuery   = errors.New("query is required")
	errQueryTooLong = fmt.Errorf("query must be at most %d characters", maxQueryLength)
)

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
				slog.Error("Error adding documents, retrying", "err", err, "start", start)
				time.Sleep(5 * time.Second)
			}
			slog.Info("Added batch", "added", added.Add(int64(end-start)), "total", len(ids))
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
		slog.Error("Error embedding query", "err", err)
		return nil, err
	}

	qr, err := c.chroma.query(ctx, c.CollectionID, embeddings[0], 10)
	if err != nil {
		slog.Error("Error querying", "err", err)
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

// validateQuery trims the query and rejects empty or overly long ones, since every query is a paid OpenAI call
func validateQuery(query string) (string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", errEmptyQuery
	}
	if utf8.RuneCountInString(query) > maxQueryLength {
		return "", errQueryTooLong
	}
	return query, nil
}

// HandleQueryRequest returns the verses closest in meaning to the query as JSON
func (c *ChromaService) HandleQueryRequest(w http.ResponseWriter, r *http.Request) {

	var queryDTO model.QueryDTO
	err := decodeJSON(w, r, &queryDTO)
	if err != nil {
		slog.Error("Error binding json", "err", err)
		writeJSON(w, http.StatusBadRequest, model.ErrorDTO{
			Error: "error binding json",
		})
		return
	}

	query, err := validateQuery(queryDTO.Query)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorDTO{
			Error: err.Error(),
		})
		return
	}

	slog.Info("Received API query request", "query", query)

	resultSlice, err := c.getQueryResults(r.Context(), query)
	if err != nil {
		slog.Error("Error getting query results", "err", err)
		writeJSON(w, http.StatusInternalServerError, model.ErrorDTO{
			Error: "error getting query results",
		})
		return
	}

	result := model.QueryResultsDTO{
		Result: *resultSlice,
	}

	writeJSON(w, http.StatusOK, result)
}

// HandleHTMXQuery returns the search results, or an error message, as an HTML fragment for htmx to swap in
func (c *ChromaService) HandleHTMXQuery(renderer *templates.Renderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		query, err := validateQuery(r.PostFormValue("query"))
		if errors.Is(err, errEmptyQuery) {
			renderer.Render(w, http.StatusOK, "results", model.SearchResultsView{})
			return
		}
		if err != nil {
			renderer.Render(w, http.StatusBadRequest, "results", model.SearchResultsView{Error: "Please shorten your search to 500 characters or fewer."})
			return
		}

		slog.Info("Received HTMX query request", "query", query)
		resultSlice, err := c.getQueryResults(r.Context(), query)
		if err != nil {
			slog.Error("Error getting query results", "err", err)
			renderer.Render(w, http.StatusInternalServerError, "results", model.SearchResultsView{Query: query, Error: "Something went wrong while searching. Please try again in a moment."})
			return
		}

		renderer.Render(w, http.StatusOK, "results", model.SearchResultsView{Query: query, Results: *resultSlice})
	}
}
