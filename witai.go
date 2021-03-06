package witai

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// DefaultVersion - https://wit.ai/docs/http/20170307
	DefaultVersion = "20170307"
)

// Client - Wit.ai client type
type Client struct {
	APIBase      string
	Token        string
	Version      string
	headerAuth   string
	headerAccept string
}

type errorResp struct {
	Body  string `json:"body"`
	Error string `json:"error"`
}

// NewClient - returns Wit.ai client for default API version
func NewClient(token string) *Client {
	return NewClientWithVersion(token, DefaultVersion)
}

// NewClientWithVersion - returns Wit.ai client for specified API version
func NewClientWithVersion(token, version string) *Client {
	headerAuth := fmt.Sprintf("Bearer %s", token)
	headerAccept := fmt.Sprintf("application/vnd.wit.%s+json", version)

	return &Client{
		APIBase:      "https://api.wit.ai",
		Token:        token,
		Version:      version,
		headerAuth:   headerAuth,
		headerAccept: headerAccept,
	}
}

func (c *Client) request(method, url string, ct string, body io.Reader) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, c.APIBase+url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", c.headerAuth)
	req.Header.Set("Accept", c.headerAccept)
	req.Header.Set("Content-Type", ct)

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		defer resp.Body.Close()

		var e *errorResp
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&e)
		if err != nil {
			return nil, errors.New("Internal Error")
		}

		// Wit.ai errors sometimes have "error", sometimes "body" message
		if len(e.Error) > 0 {
			return nil, errors.New(e.Error)
		}

		return nil, errors.New(e.Body)
	}

	return resp.Body, nil
}
