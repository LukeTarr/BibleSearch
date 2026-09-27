package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	chromaTenant   = "default_tenant"
	chromaDatabase = "default_database"
)

// chromaClient is a minimal client for the Chroma v2 REST API, covering only what this app uses
type chromaClient struct {
	baseURL string
	http    *http.Client
}

func newChromaClient(chromaURL string) *chromaClient {
	return &chromaClient{
		baseURL: strings.TrimSuffix(chromaURL, "/") + "/api/v2",
		http:    &http.Client{Timeout: 2 * time.Minute},
	}
}

// verseMetadata is the metadata stored with every verse
type verseMetadata struct {
	Book    string `json:"book"`
	Chapter string `json:"chapter"`
	Verse   string `json:"verse"`
}

type chromaQueryResult struct {
	IDs       [][]string        `json:"ids"`
	Documents [][]string        `json:"documents"`
	Metadatas [][]verseMetadata `json:"metadatas"`
	Distances [][]float64       `json:"distances"`
}

func (c *chromaClient) do(ctx context.Context, method string, path string, reqBody any, respOut any) error {
	var body io.Reader
	if reqBody != nil {
		b, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chroma %s %s: %s: %s", method, path, resp.Status, msg)
	}

	if respOut == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(respOut)
}

func collectionsPath() string {
	return "/tenants/" + chromaTenant + "/databases/" + chromaDatabase + "/collections"
}

func collectionPath(collectionID string) string {
	return collectionsPath() + "/" + url.PathEscape(collectionID)
}

func (c *chromaClient) reset(ctx context.Context) error {
	return c.do(ctx, http.MethodPost, "/reset", nil, nil)
}

// getOrCreateCollection returns the collection's ID. Chroma's default distance is L2.
func (c *chromaClient) getOrCreateCollection(ctx context.Context, name string) (string, error) {
	req := map[string]any{
		"name":          name,
		"get_or_create": true,
	}
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.do(ctx, http.MethodPost, collectionsPath(), req, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *chromaClient) add(ctx context.Context, collectionID string, ids []string, embeddings [][]float32, documents []string, metadatas []verseMetadata) error {
	req := map[string]any{
		"ids":        ids,
		"embeddings": embeddings,
		"documents":  documents,
		"metadatas":  metadatas,
	}
	return c.do(ctx, http.MethodPost, collectionPath(collectionID)+"/add", req, nil)
}

func (c *chromaClient) query(ctx context.Context, collectionID string, embedding []float32, n int) (*chromaQueryResult, error) {
	req := map[string]any{
		"query_embeddings": [][]float32{embedding},
		"n_results":        n,
		"include":          []string{"documents", "metadatas", "distances"},
	}
	var resp chromaQueryResult
	if err := c.do(ctx, http.MethodPost, collectionPath(collectionID)+"/query", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *chromaClient) count(ctx context.Context, collectionID string) (int, error) {
	var n int
	if err := c.do(ctx, http.MethodGet, collectionPath(collectionID)+"/count", nil, &n); err != nil {
		return 0, err
	}
	return n, nil
}
