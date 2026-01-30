package semantic

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)


type Client struct {
	baseURL string 
	client *http.Client
}


func New(baseURL string) *Client{
	return &Client{
		baseURL:	baseURL,
		client:		&http.Client{
			Timeout:    2 * time.Second,
		},
	}
}


type request struct {
	Word	string	`json:"word"`
	Target	string	`json:"target"`
}


type response struct {
	Similarity	float64  `json:"similarity"`
}


func (c *Client) Similarity(word, target string) (float64, error) {
	reqBody, _  := json.Marshal(request{Word: word, Target: target})

	resp, err := c.client.Post(
		c.baseURL+"/similarity",
		"application/json",
		bytes.NewBuffer(reqBody),
	)

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

