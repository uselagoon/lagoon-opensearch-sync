package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
)

// Mapping represents a field mapping.
type Mapping struct {
	Type            string `json:"type"`
	IgnoreMalformed bool   `json:"ignore_malformed"`
}

// DynamicTemplate represents a dynamic template
type DynamicTemplate struct {
	MatchMappingType string  `json:"match_mapping_type,omitempty"`
	MatchPattern     string  `json:"match_pattern,omitempty"`
	Match            string  `json:"match"`
	Mapping          Mapping `json:"mapping"`
}

// Mappings represents Opensearch index mappings.
type Mappings struct {
	DynamicTemplates []map[string]DynamicTemplate `json:"dynamic_templates"`
}

// Template represents an Opensearch template.
type Template struct {
	Mappings *Mappings `json:"mappings"`
}

// IndexTemplate represents an Opensearch index template.
type IndexTemplate struct {
	Name                    string                  `json:"name"`
	IndexTemplateDefinition IndexTemplateDefinition `json:"index_template"`
}

// IndexTemplateDefinition contain only the definition of the IndexTemplate
// (excluding the name). This type, which is embedded in IndexTemplate, exists
// so that a valid PUT request can be easily made to the Opensearch API. This
// requires omitting the Name field.
type IndexTemplateDefinition struct {
	ComposedOf    []string `json:"composed_of,omitempty"`
	IndexPatterns []string `json:"index_patterns"`
	Template      Template `json:"template"`
}

// IndexTemplatesSlice is used only for unmarshalling the JSON data returned by
// the Opensearch index templates API.
type IndexTemplatesSlice struct {
	IndexTemplates []IndexTemplate `json:"index_templates"`
}

// indexTemplatesMap unmarshals the data returned from the Opensearch index
// templates API and returns it as a map of names to IndexTemplate objects.
func indexTemplatesMap(its *IndexTemplatesSlice) map[string]IndexTemplate {
	itm := map[string]IndexTemplate{}
	for _, t := range its.IndexTemplates {
		itm[t.Name] = t
	}
	return itm
}

// RawIndexTemplates returns the raw JSON index templates representation from
// the Opensearch API.
func (c *Client) RawIndexTemplates(ctx context.Context) ([]byte, error) {
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
	ctx context.Context) (map[string]IndexTemplate, error) {
	data, err := c.RawIndexTemplates(ctx)
	if err != nil {
		return nil,
			fmt.Errorf("couldn't get index templates from Opensearch API: %v", err)
	}
	var its IndexTemplatesSlice
	if err := json.Unmarshal(data, &its); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal index templates: %v", err)
	}
	return indexTemplatesMap(&its), nil
}

// CreateIndexTemplate creates the given index template in Opensearch.
func (c *Client) CreateIndexTemplate(ctx context.Context, name string,
	it *IndexTemplate) error {
	// Marshal payload. Payload only consists of IndexTemplateRules because the
	// name field is not writable.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(it.IndexTemplateDefinition); err != nil {
		return fmt.Errorf("couldn't marshal index template: %v", err)
	}
	// construct request
	url := *c.baseURL
	url.Path = path.Join(c.baseURL.Path,
		"/_index_template/", name)
	req, err := http.NewRequestWithContext(ctx, "PUT", url.String(), &buf)
	if err != nil {
		return fmt.Errorf("couldn't construct create index template request: %v",
			err)
	}
	req.Header.Set("Content-Type", "application/json")
	// make request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't create index template: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bad create index template response: %d\n%s",
			res.StatusCode, body)
	}
	return nil
}

// DeleteIndexTemplate deletes the named index template from Opensearch.
func (c *Client) DeleteIndexTemplate(ctx context.Context, name string) error {
	// construct request
	url := *c.baseURL
	url.Path = path.Join(c.baseURL.Path,
		"/_index_template/", name)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url.String(), nil)
	if err != nil {
		return fmt.Errorf("couldn't construct delete index template request: %v",
			err)
	}
	// make request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't delete index template: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bad delete index template response: %d\n%s",
			res.StatusCode, body)
	}
	return nil
}
