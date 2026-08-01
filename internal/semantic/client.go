package semantic

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Semantic is the interface for computing word similarity.
// Both the HTTP Client (legacy) and the future LocalClient implement this.
type Semantic interface {
	Similarity(ctx context.Context, word, target string) (float64, error)
}

// Client calls the Python FastAPI semantic service over HTTP.
// Kept for compatibility; swap out for LocalClient once embeddings are pre-computed.
type Client struct {
	baseURL string
	client  *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 2 * time.Second},
	}
}

type request struct {
	Word   string `json:"word"`
	Target string `json:"target"`
}

type response struct {
	Similarity float64 `json:"similarity"`
}

// Similarity implements the Semantic interface.
// Uses context so in-flight requests can be cancelled (e.g. when a game ends).
func (c *Client) Similarity(ctx context.Context, word, target string) (float64, error) {
	reqBody, _ := json.Marshal(request{Word: word, Target: target})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/similarity", bytes.NewBuffer(reqBody))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var res response
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, err
	}
	return res.Similarity, nil
}
