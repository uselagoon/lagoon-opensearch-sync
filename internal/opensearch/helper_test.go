package opensearch

import (
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

// this test helper facilitates unit testing of private functions.

var (
	IndexTemplatesMap  = indexTemplatesMap
	ParseIndexPatterns = parseIndexPatterns
)

// NewTestClient creates a new Opensearch client for testing.
func NewTestClient(
	baseURLRaw string,
	searchSize uint,
) (*Client, error) {
	baseURL, err := url.Parse(baseURLRaw)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse URL: %v", err)
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: http.DefaultClient,
		log:        zap.Must(zap.NewDevelopment()),
		searchSize: searchSize,
	}, nil
}
