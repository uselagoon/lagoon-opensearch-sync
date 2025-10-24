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
	"strings"

	"go.uber.org/zap"
)

var (
	// indexName matches the raw name of an index-pattern index name and its
	// migration number
	indexName = regexp.MustCompile(`^\.kibana(?:_(.+))?_([0-9]+)$`)
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
	Sort   []int  `json:"sort"`
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
	SearchAfter []int                        `json:"search_after,omitempty"`
	Size        uint                         `json:"size"`
	Sort        map[string]map[string]string `json:"sort"`
}

// newSearchBody returns an Opensearch search request body.
//
// searchSize populates the size field, and controls the number of results
// returned. The Maximum value accepted by the Opensearch API is 10000.
//
// searchAfter populates the search_after field, and allows paging through
// results.
func newSearchBody(searchSize uint, searchAfter []int) (*bytes.Buffer, error) {
	body := SearchBody{
		Query: SearchQuery{
			Term: map[string]map[string]string{
				"type": {
					"value": "index-pattern",
				},
			},
		},
		SearchAfter: searchAfter,
		Size:        searchSize,
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
// Opensearch API. The searchAfter parameter allows specifying a search_after
// date.
//
// If searchAfter is empty or nil, search_after is omitted from the Opensearch
// API request, and results will be returned from the "first page".
//
// https://docs.opensearch.org/latest/search-plugins/searching-data/paginate/
func (c *Client) RawIndexPatterns(
	ctx context.Context,
	searchSize uint,
	searchAfter []int,
) ([]byte, error) {
	buf, err := newSearchBody(searchSize, searchAfter)
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

// parseIndexName takes a raw index name with the ".kibana_" prefix and "_n"
// suffix (where "n" is the migration number). It returns the index name
// stripped of the prefix and suffix, the migration number as an int, and an
// error (if any).
func parseIndexName(rawIndex string) (string, int, error) {
	matches := indexName.FindStringSubmatch(rawIndex)
	if len(matches) != 3 {
		return "", 0, fmt.Errorf("invalid index name: %s", rawIndex)
	}
	var index string
	if matches[1] == "" {
		index = "global_tenant"
	} else {
		index = matches[1]
	}
	migration, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", 0, fmt.Errorf("couldn't parse migration number: %v", err)
	}
	if migration < 1 {
		return "", 0, fmt.Errorf("invalid migration number: %d", migration)
	}
	return index, migration, nil
}

// indexMaxMigration iterates over hits and returns a map containing the unique
// index names found, mapped to the maximum migration number of each of those
// indices. The index names are stripped of their ".kibana_" prefix and their
// "_n" suffix, where "n" is the migration number.
func indexMaxMigration(hits []IndexPattern) (map[string]int, error) {
	maxMigration := map[string]int{}
	for _, hit := range hits {
		index, migration, err := parseIndexName(hit.Index)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse index name %s: %v", hit.Index, err)
		}
		if maxMigration[index] < migration {
			maxMigration[index] = migration
		}
	}
	return maxMigration, nil
}

// parseIndexPatterns takes a slice of IndexPattern query results.
//
// It returns a map of Index Names (corresponding to tenants) to a map of Index
// Pattern titles to a slice of corresponding Index Pattern IDs.
func parseIndexPatterns(
	hits []IndexPattern,
) (map[string]map[string][]string, error) {
	// initialise the results map
	var indexPatterns = map[string]map[string][]string{}
	// handle the case of zero index patterns
	if len(hits) == 0 {
		return indexPatterns, nil
	}
	maxMigration, err := indexMaxMigration(hits)
	if err != nil {
		return nil, fmt.Errorf("couldn't get max migrations: %v", err)
	}
	for _, hit := range hits {
		indexName, migration, err := parseIndexName(hit.Index)
		if err != nil {
			return nil, fmt.Errorf("couldn't parse index name %s: %v", hit.Index, err)
		}
		if maxMigration[indexName] != migration {
			// ignore old migrations of indices
			continue
		}
		// initialize the nested map
		if indexPatterns[indexName] == nil {
			indexPatterns[indexName] = map[string][]string{}
		}
		// search results prefix ID with "index-pattern:", which is stripped here
		// because the prefix is not used when referring to the index pattern by ID
		// in other API requests.
		patternID := strings.TrimPrefix(hit.ID, "index-pattern:")
		// Multiple identically named index patterns may be added to a single
		// tenant, so map the index pattern names to a slice of IDs.
		indexPatterns[indexName][hit.Source.IndexPattern.Title] =
			append(indexPatterns[indexName][hit.Source.IndexPattern.Title], patternID)
	}
	return indexPatterns, nil
}

// IndexPatterns returns all Opensearch index patterns as a map of index names
// (which are derived from tenant names) to map of index pattern titles to
// index pattern IDs, which is set if the index pattern exists in the tenant.
func (c *Client) IndexPatterns(ctx context.Context) (
	map[string]map[string][]string, error) {
	var searchAfter []int
	var s SearchResult
	var rawResults []IndexPattern
	for {
		data, err := c.RawIndexPatterns(ctx, c.searchSize, searchAfter)
		if err != nil {
			return nil,
				fmt.Errorf("couldn't get index patterns from Opensearch API: %v", err)
		}
		// unpack index patterns in query result
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf(
				"couldn't unmarshal index patterns search result: %v", err)
		}
		rawResults = append(rawResults, s.Hits.Hits...)
		// unpack index patterns in query result
		if len(s.Hits.Hits) < int(c.searchSize) {
			c.log.Debug("got all index patterns, parsing result",
				zap.Int("hits", len(rawResults)))
			// we have got all the index patterns...
			return parseIndexPatterns(rawResults)
		}
		// ...otherwise we need to do another request
		c.log.Debug("partial index pattern search response: scrolling results")
		searchAfter = s.Hits.Hits[len(s.Hits.Hits)-1].Sort
	}
}
