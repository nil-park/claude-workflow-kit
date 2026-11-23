package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
)

const (
	embeddingDimensions = 768
	embeddingModel      = "gemini-embedding-2"
	cloudPlatformScope  = "https://www.googleapis.com/auth/cloud-platform"
	vertexEndpointFmt   = "https://aiplatform.googleapis.com/v1/projects/%s/locations/global/publishers/google/models/%s:embedContent"
	taskTypeDocument    = "RETRIEVAL_DOCUMENT"
	taskTypeQuery       = "RETRIEVAL_QUERY"
	embedHTTPTimeout    = 20 * time.Second
	maxEmbedRespBytes   = 4 << 20
	shortIDAlphabet     = "abcdefghijklmnopqrstuvwxyz234567"
	shortIDLength       = 10
)

func defaultGCPCredentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".config", "mem4j", ".secrets", "gcp.json"), nil
}

func generateShortID() (string, error) {
	var raw [7]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate short id: %w", err)
	}
	var bits uint64
	for _, b := range raw {
		bits = (bits << 8) | uint64(b)
	}
	out := make([]byte, shortIDLength)
	for i := range out {
		out[i] = shortIDAlphabet[bits&31]
		bits >>= 5
	}
	return string(out), nil
}

type vertexEmbedder struct {
	endpoint string
	client   *http.Client
}

func newVertexEmbedder(ctx context.Context, credPath string) (*vertexEmbedder, error) {
	if credPath == "" {
		var err error
		credPath, err = defaultGCPCredentialsPath()
		if err != nil {
			return nil, err
		}
	}
	raw, err := os.ReadFile(credPath)
	if err != nil {
		return nil, fmt.Errorf("read GCP credentials %s: %w", credPath, err)
	}
	var meta struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil, fmt.Errorf("parse GCP credentials %s: %w", credPath, err)
	}
	if strings.TrimSpace(meta.ProjectID) == "" {
		return nil, fmt.Errorf("%s: project_id is missing", credPath)
	}
	cfg, err := google.JWTConfigFromJSON(raw, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("load GCP service account %s: %w", credPath, err)
	}
	client := cfg.Client(ctx)
	client.Timeout = embedHTTPTimeout
	return &vertexEmbedder{
		endpoint: fmt.Sprintf(vertexEndpointFmt, meta.ProjectID, embeddingModel),
		client:   client,
	}, nil
}

type embedRequest struct {
	Content              embedContent `json:"content"`
	TaskType             string       `json:"taskType,omitempty"`
	OutputDimensionality int          `json:"outputDimensionality"`
}

type embedContent struct {
	Parts []embedPart `json:"parts"`
}

type embedPart struct {
	Text string `json:"text"`
}

type embedResponse struct {
	Embedding struct {
		Values []float64 `json:"values"`
	} `json:"embedding"`
}

func (v *vertexEmbedder) Embed(ctx context.Context, text, taskType string) ([]float64, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.New("embedding input text cannot be empty")
	}
	payload, err := json.Marshal(embedRequest{
		Content:              embedContent{Parts: []embedPart{{Text: text}}},
		TaskType:             taskType,
		OutputDimensionality: embeddingDimensions,
	})
	if err != nil {
		return nil, fmt.Errorf("encode embed request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call Vertex AI embedContent: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxEmbedRespBytes))
	if err != nil {
		return nil, fmt.Errorf("read embed response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Vertex AI embedContent returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var parsed embedResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	if len(parsed.Embedding.Values) != embeddingDimensions {
		return nil, fmt.Errorf("expected %d embedding dimensions, got %d", embeddingDimensions, len(parsed.Embedding.Values))
	}
	return parsed.Embedding.Values, nil
}
