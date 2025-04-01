package dashboards

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
)

// createIndexPatternRequest represents an Opensearch Dashboards index pattern
// create request body.
type createIndexPatternRequest struct {
	IndexPattern struct {
		Title         string `json:"title"`
		TimeFieldName string `json:"timeFieldName"`
	} `json:"attributes"`
}

// CreateIndexPattern creates the given index pattern in the given tenant in
// Opensearch Dashboards.
func (c *Client) CreateIndexPattern(ctx context.Context,
	tenant, pattern string) error {
	// marshal body
	cipReq := createIndexPatternRequest{}
	cipReq.IndexPattern.TimeFieldName = "@timestamp"
	cipReq.IndexPattern.Title = pattern
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(cipReq); err != nil {
		return fmt.Errorf("couldn't marshal request body: %v", err)
	}
	// construct request
	url := *c.baseURL
	url.Path = path.Join(c.baseURL.Path,
		"/api/saved_objects/index-pattern", pattern)
	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), &buf)
	if err != nil {
		return fmt.Errorf("couldn't construct request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("osd-xsrf", "true")
	switch tenant {
	case "global_tenant":
		// the global tenant is special cased in the dashboards security plugin
		//
		// https://github.com/opensearch-project/security-dashboards-plugin/blob/
		// 	5e99582ea55176a16f0190998abcc99c99e7f974/server/multitenancy/
		//  tenant_resolver.ts#L23
		req.Header.Set("securitytenant", "global")
	default:
		req.Header.Set("securitytenant", tenant)
	}
	// make request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %v", err)
	}
	defer res.Body.Close() // nolint: errcheck
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf(
			"bad response creating index pattern %s for tenant %s: %d\n%s",
			pattern, tenant, res.StatusCode, body)
	}
	return nil
}

// DeleteIndexPattern deletes the named index pattern from the given tenant in
// Opensearch Dashboards.
func (c *Client) DeleteIndexPattern(ctx context.Context,
	tenant, pattern string) error {
	// construct request
	url := *c.baseURL
	url.Path = path.Join(c.baseURL.Path,
		"/api/saved_objects/index-pattern", pattern)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url.String(), nil)
	if err != nil {
		return fmt.Errorf("couldn't construct delete request: %v", err)
	}
	req.Header.Set("osd-xsrf", "true")
	switch tenant {
	case "global_tenant":
		// the global tenant is special cased in the dashboards security plugin
		//
		// https://github.com/opensearch-project/security-dashboards-plugin/blob/
		// 	5e99582ea55176a16f0190998abcc99c99e7f974/server/multitenancy/
		//  tenant_resolver.ts#L23
		req.Header.Set("securitytenant", "global")
	default:
		req.Header.Set("securitytenant", tenant)
	}
	// make request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %v", err)
	}
	defer res.Body.Close() // nolint: errcheck
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bad response: %d\n%s", res.StatusCode, body)
	}
	return nil
}
