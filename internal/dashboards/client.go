// Package dashboards implements an Opensearch Dashboards API client.
package dashboards

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client is an Opensearch Dashboards client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient creates a new Opensearch Dashboards client.
func NewClient(baseURL, username, password string, timeout time.Duration) (*Client, error) {
	// parse URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse base URL %s: %v", baseURL, err)
	}
	// construct client
	return &Client{
		baseURL:    u,
		httpClient: httpClient(username, password, timeout),
	}, nil
}
