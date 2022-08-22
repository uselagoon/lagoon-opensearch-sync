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
func (c *Client) Tenants(ctx context.Context) (map[string]Tenant, error) {
	rawTenants, err := c.rawTenants(ctx)
	if err != nil {
		return nil,
			fmt.Errorf("couldn't get tenants from Opensearch API: %v", err)
	}
	var t map[string]Tenant
	return t, json.Unmarshal(rawTenants, &t)
}

// CreateTenant creates the given tenant in Opensearch.
func (c *Client) CreateTenant(ctx context.Context, name string,
	tenant *Tenant) error {
	// marshal payload
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(tenant); err != nil {
		return fmt.Errorf("couldn't marshal tenant: %v", err)
	}
	// construct request
	url := *c.baseURL
	url.Path = path.Join(c.baseURL.Path,
		"/_plugins/_security/api/tenants/", name)
	req, err := http.NewRequestWithContext(ctx, "PUT", url.String(), &buf)
	if err != nil {
		return fmt.Errorf("couldn't construct create tenant request: %v", err)
	}
	// make request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't create tenant: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bad create tenant response: %d\n%s", res.StatusCode,
			body)
	}
	return nil
}

// DeleteTenant deletes the named tenant from Opensearch.
func (c *Client) DeleteTenant(ctx context.Context, name string) error {
	// construct request
	url := *c.baseURL
	url.Path = path.Join(c.baseURL.Path,
		"/_plugins/_security/api/tenants/", name)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url.String(), nil)
	if err != nil {
		return fmt.Errorf("couldn't construct delete tenant request: %v", err)
	}
	// make request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't delete tenant: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bad delete tenant response: %d\n%s", res.StatusCode,
			body)
	}
	return nil
}
