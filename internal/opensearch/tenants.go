package opensearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/davecgh/go-spew/spew"
)

// Tenant represents an Opensearch Tenant.
type Tenant struct {
	Description string `json:"description"`
	Hidden      bool   `json:"hidden"`
	Reserved    bool   `json:"reserved"`
	Static      bool   `json:"static"`
}

// rawTenants returns the raw JSON tenants representation from the
// Opensearch API.
func (c *Client) rawTenants(ctx context.Context) ([]byte, error) {
	tenantsURL := *c.baseURL
	tenantsURL.Path = path.Join(c.baseURL.Path,
		"/_plugins/_security/api/tenants/")
	req, err := http.NewRequestWithContext(ctx, "GET", tenantsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't construct tenants request: %v", err)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't get tenants: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bad tenants response: %d\n%s",
			res.StatusCode, body)
	}
	return io.ReadAll(res.Body)
}

// Tenants returns all Opensearch Tenants.
func (c *Client) Tenants(
	ctx context.Context) (map[string]RoleMapping, error) {
	rawTenants, err := c.rawTenants(ctx)
	if err != nil {
		return nil,
			fmt.Errorf("couldn't get tenants from Opensearch API: %v", err)
	}
	spew.Dump(string(rawTenants))
	var rm map[string]RoleMapping
	return rm, json.Unmarshal(rawTenants, &rm)
}
