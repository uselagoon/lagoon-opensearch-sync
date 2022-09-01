package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// maximum size of search results returned by Opensearch
// https://opensearch.org/docs/latest/opensearch/ux/#scroll-search
const searchSize = 10000

// Source represents the source field in an Opensearch search result.
type Source struct {
	UpdatedAt string `json:"updated_at"`
}

// IndexPattern represents an Opensearch Dashboards index pattern.
type IndexPattern struct {
	ID     string `json:"_id"`
	Index  string `json:"_index"`
	Source Source `json:"_source"`
}

// SearchHits represents the array of hits in a search result.
type SearchHits struct {
	Hits []IndexPattern `json:"hits"`
}

// SearchResult represents the result of an Opensearch search.
type SearchResult struct {
	Hits SearchHits `json:"hits"`
}

// SearchQuery represents the query field of an Opensearch search request.
type SearchQuery struct {
	Term map[string]map[string]string `json:"term"`
}

// SearchBody represents the body of an Opensearch search request.
type SearchBody struct {
	Query       SearchQuery                  `json:"query"`
	SearchAfter []string                     `json:"search_after,omitempty"`
	Size        uint                         `json:"size"`
	Sort        map[string]map[string]string `json:"sort"`
}

// newSearchBody returns an Opensearch search request body.
// If after is given it populates the search_after field.
func newSearchBody(after string) (*bytes.Buffer, error) {
	var searchAfter []string
	if len(after) > 0 {
		searchAfter = append(searchAfter, after)
	}
	body := SearchBody{
		Query: SearchQuery{
			Term: map[string]map[string]string{
				"type": {
					"value": "index-pattern",
				},
			},
		},
		SearchAfter: searchAfter,
		Size:        10000, // Opensearch query max
		Sort: map[string]map[string]string{
			"updated_at": {
				"order": "asc",
			},
		},
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	return &buf, enc.Encode(&body)
}

// RawIndexPatterns returns the raw JSON index patterns representation from the
// Opensearch API. The after parameter allows specifying a search_after date.
// If after is an empty string, search_after is omitted from the Opensearch API
// request.
// https://opensearch.org/docs/latest/opensearch/ux/#paginate-results
func (c *Client) RawIndexPatterns(ctx context.Context,
	after string) ([]byte, error) {
	buf, err := newSearchBody(after)
	if err != nil {
		return nil, fmt.Errorf("couldn't construct search body: %v", err)
	}
	indexPatternsURL := *c.baseURL
	indexPatternsURL.Path = path.Join(c.baseURL.Path, ".kibana*/_search")
	req, err := http.NewRequestWithContext(ctx, "GET", indexPatternsURL.String(),
		buf)
	if err != nil {
		return nil, fmt.Errorf("couldn't construct indexPatterns request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	q := req.URL.Query()
	q.Add("q", "type:index-pattern")
	q.Add("size", strconv.Itoa(searchSize))
	req.URL.RawQuery = q.Encode()
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't get indexPatterns: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bad indexPatterns response: %d\n%s",
			res.StatusCode, body)
	}
	return io.ReadAll(res.Body)
}

// IndexPatterns returns all Opensearch index patterns.
func (c *Client) IndexPatterns(ctx context.Context) (
	map[string]map[string]bool, error) {
	indexPatterns := map[string]map[string]bool{}
	var after, index string
	for {
		rawIndexPatterns, err := c.RawIndexPatterns(ctx, after)
		if err != nil {
			return nil,
				fmt.Errorf("couldn't get index patterns from Opensearch API: %v", err)
		}
		// unpack all index patterns
		var s SearchResult
		if err = json.Unmarshal(rawIndexPatterns, &s); err != nil {
			return nil, fmt.Errorf(
				"couldn't unmarshal index patterns search result: %v", err)
		}
		for _, hit := range s.Hits.Hits {
			if hit.Index == ".kibana_1" {
				index = "global_tenant"
			} else {
				index = strings.TrimPrefix(hit.Index, ".kibana_")
				index = strings.TrimSuffix(index, "_1")
				// sanity-check the index pattern format and return an error if it is
				// not as expected.
				if len(strings.Split(index, "_")) != 2 {
					return nil, fmt.Errorf("unexpected index name: %v", hit.Index)
				}
			}
			// initialize the nested map
			if indexPatterns[index] == nil {
				indexPatterns[index] = map[string]bool{}
			}
			indexPatterns[index][strings.TrimPrefix(hit.ID, "index-pattern:")] = true
		}
		if len(s.Hits.Hits) < searchSize {
			c.log.Debug("got all index patterns, returning result",
				zap.Int("hits", len(s.Hits.Hits)))
			break // we have got all the index patterns...
		}
		// ...otherwise we need to do another request
		c.log.Debug("partial index pattern search response: scrolling results")
		after = s.Hits.Hits[searchSize-1].Source.UpdatedAt
	}
	return indexPatterns, nil
}
