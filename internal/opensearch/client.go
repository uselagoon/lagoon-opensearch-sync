package opensearch

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Client is an Opensearch client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new Opensearch client.
func NewClient(ctx context.Context, baseURL, username,
	password string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse base URL %s: %v", baseURL, err)
	}
	return &Client{
		baseURL:    u,
		httpClient: httpClient(username, password),
	}, nil
}


