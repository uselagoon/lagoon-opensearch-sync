package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strconv"

	"go.uber.org/zap"
)

// maximum size of search results returned by Opensearch
// https://opensearch.org/docs/latest/opensearch/ux/#scroll-search
const searchSize = 10000

var (
	// globalTenantIndexName matches the name of the index that the global tenant
	// index patterns are stored in.
	globalTenantIndexName = regexp.MustCompile(`^\.kibana_[0-9]+$`)
	// tenantIndexName matches the name of the index that regular tenant index
	// patterns are stored in. The format of the match is
	// <hashInt>_<sanitizedName>.
	tenantIndexName = regexp.MustCompile(`^\.kibana_(.+)_[0-9]+$`)
)

// SourceIndexPattern represents the index pattern definition inside the
// index-pattern index.
type SourceIndexPattern struct {
	Title string `json:"title"`
}

// Source represents the source field in an Opensearch search result.
type Source struct {
	UpdatedAt    string             `json:"updated_at"`
	IndexPattern SourceIndexPattern `json:"index-pattern"`
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

// IndexPatterns returns all Opensearch index patterns as a map of index names
// (which are derived from tenant names) to map of index pattern titles to
// bool, which is set to true if the index pattern exists in the tenant.
//
// This function ignores migrated .kibana indices, so it may set the same index
// pattern to true in the map more than once if e.g. indices named
// .kibana_mytenant_{1,2,3} all exist. TODO: figure out how to tell which of
// these indices represents the current index-pattern.
func (c *Client) IndexPatterns(ctx context.Context) (
	map[string]map[string]bool, error) {
	indexPatterns := map[string]map[string]bool{}
	var after string
	for {
		rawIndexPatterns, err := c.RawIndexPatterns(ctx, after)
		if err != nil {
			return nil,
				fmt.Errorf("couldn't get index patterns from Opensearch API: %v", err)
		}
		searchResultSize, lastUpdatedAt, err :=
			parseIndexPatterns(rawIndexPatterns, indexPatterns)
		if err != nil {
			return nil,
				fmt.Errorf("couldn't parse index patterns: %v", err)
		}
		if searchResultSize < searchSize {
			c.log.Debug("got all index patterns, returning result",
				zap.Int("hits", searchResultSize))
			break // we have got all the index patterns...
		}
		// ...otherwise we need to do another request
		c.log.Debug("partial index pattern search response: scrolling results")
		after = lastUpdatedAt
	}
	return indexPatterns, nil
}

// parseIndexPatterns takes the raw index patterns search results as a JSON
// blob, and a map to store results.
// It fills out the map according to the index patterns that it finds, and
// returns the number of search results found, the updated at date on the last
// search result, and an error (if any).
func parseIndexPatterns(data []byte,
	indexPatterns map[string]map[string]bool) (int, string, error) {
	// unpack all index patterns
	var s SearchResult
	var index string
	if err := json.Unmarshal(data, &s); err != nil {
		return 0, "", fmt.Errorf(
			"couldn't unmarshal index patterns search result: %v", err)
	}
	// handle the case of zero index patterns
	if len(s.Hits.Hits) == 0 {
		return 0, "1970-01-01T00:00:00Z", nil
	}
	for _, hit := range s.Hits.Hits {
		if globalTenantIndexName.MatchString(hit.Index) {
			index = "global_tenant"
		} else {
			matches := tenantIndexName.FindStringSubmatch(hit.Index)
			// sanity-check the index pattern format and return an error if it is
			// not as expected.
			if len(matches) != 2 {
				return 0, "", fmt.Errorf("unexpected index name: %v", hit.Index)
			}
			index = matches[1]
		}
		// initialize the nested map
		if indexPatterns[index] == nil {
			indexPatterns[index] = map[string]bool{}
		}
		indexPatterns[index][hit.Source.IndexPattern.Title] = true
	}
	return len(s.Hits.Hits), s.Hits.Hits[len(s.Hits.Hits)-1].Source.UpdatedAt, nil
}
