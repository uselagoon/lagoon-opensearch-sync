// Package opensearch implements a client for interacting with Opensearch.
package opensearch

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// Maximum size of search results returned by Opensearch.
// https://docs.opensearch.org/latest/search-plugins/searching-data/paginate/
const searchSizeMax = 10000

// Client is an Opensearch client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	log        *zap.Logger
	searchSize uint
}

// NewClient creates a new Opensearch client.
func NewClient(
	log *zap.Logger,
	baseURL,
	username,
	password,
	caCertificate string,
	timeout time.Duration,
) (*Client, error) {
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
		httpClient: httpClient(username, password, ca, timeout),
		log:        log,
		searchSize: searchSizeMax,
	}, nil
}
