package opensearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
)

// Alias represents an Opensearch index alias.
type Alias struct{}

// Settings represents Opensearch index settings.
// It handles a subset of all index settings.
type Settings struct {
	NumberOfShards   int `json:"number_of_shards"`
	NumberOfReplicas int `json:"number_of_replicas"`
}

// Mapping represents a field mapping.
type Mapping struct {
	Type  string `json:"type"`
	Index *bool  `json:"index"`
}

// Mappings represents Opensearch index mappings.
type Mappings struct {
	Properties map[string]Mapping `json:"properties"`
}

// Template represents an Opensearch template.
type Template struct {
	Aliases  map[string]Alias `json:"aliases"`
	Settings *Settings        `json:"settings"`
	Mappings *Mappings        `json:"mappings"`
}

// IndexTemplate represents an Opensearch index template.
type IndexTemplate struct {
	IndexPatterns []string `json:"index_patterns"`
	Template      Template `json:"template"`
}

// rawIndexTemplates returns the raw JSON index templates representation from
// the Opensearch API.
func (c *Client) rawIndexTemplates(ctx context.Context) ([]byte, error) {
	url := *c.baseURL
	url.Path = path.Join(c.baseURL.Path,
		"/_index_template/")
	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't construct index template request: %v", err)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't get index template: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bad index template response: %d\n%s",
			res.StatusCode, body)
	}
	return io.ReadAll(res.Body)
}

// IndexTemplates returns all Opensearch IndexTemplates.
func (c *Client) IndexTemplates(
	ctx context.Context) ([]IndexTemplate, error) {
	data, err := c.rawIndexTemplates(ctx)
	if err != nil {
		return nil,
			fmt.Errorf("couldn't get index templates from Opensearch API: %v", err)
	}
	var it struct {
		IndexTemplates []IndexTemplate `json:"index_templates"`
	}
	return it.IndexTemplates, json.Unmarshal(data, &it)
}
