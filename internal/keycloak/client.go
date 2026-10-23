// Package keycloak implements a keycloak client for Lagoon.
package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client is a Keycloak admin client.
type Client struct {
	baseURL        *url.URL
	httpClient     *http.Client
	groupsPageSize uint
}

// NewClientCredentialsClient creates a new keycloak client.
//
// clientTimeout controls the HTTP client request timeout used for all
// Keycloak API requests (OIDC discovery and Admin API calls).
//
// groupsPageSize controls the number of groups requested per page when
// paginating through the Keycloak Admin API groups endpoint.
func NewClientCredentialsClient(
	ctx context.Context,
	baseURL, clientID, clientSecret string,
	clientTimeout time.Duration,
	groupsPageSize uint,
) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse base URL %s: %v", baseURL, err)
	}
	httpClient, err := httpClient(ctx, *u, "lagoon", clientID, clientSecret,
		clientTimeout)
	if err != nil {
		return nil, fmt.Errorf("couldn't get keycloak http client: %v", err)
	}
	return &Client{
		baseURL:        u,
		httpClient:     httpClient,
		groupsPageSize: groupsPageSize,
	}, nil
}
