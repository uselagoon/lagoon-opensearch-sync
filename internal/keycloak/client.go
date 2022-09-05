// Package keycloak implements a keycloak client for Lagoon.
package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Client is a Keycloak admin client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClientCredentialsClient creates a new keycloak client.
func NewClientCredentialsClient(ctx context.Context, baseURL, clientID,
	clientSecret string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse base URL %s: %v", baseURL, err)
	}
	httpClient, err := httpClient(ctx, *u, "lagoon", clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("couldn't get keycloak http client: %v", err)
	}
	return &Client{
		baseURL:    u,
		httpClient: httpClient,
	}, nil
}
