package opensearch

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

// Client is an Opensearch client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	log        *zap.Logger
}

// NewClient creates a new Opensearch client.
func NewClient(ctx context.Context, log *zap.Logger, baseURL, username,
	password, caCertificate string) (*Client, error) {
	// parse URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse base URL %s: %v", baseURL, err)
	}
	// parse certificate
	block, _ := pem.Decode([]byte(caCertificate))
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("couldn't decode CA certificate: %v", err)
	}
	ca, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse CA certificate: %v", err)
	}
	// construct client
	return &Client{
		baseURL:    u,
		httpClient: httpClient(username, password, ca),
		log:        log,
	}, nil
}
