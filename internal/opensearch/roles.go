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

// TenantPermission represents an Opensearch tenant permission.
type TenantPermission struct {
	AllowedActions []string `json:"allowed_actions"`
	TenantPatterns []string `json:"tenant_patterns"`
}

// IndexPermission represents an Opensearch index permission.
type IndexPermission struct {
	AllowedActions []string `json:"allowed_actions"`
	FLS            []string `json:"fls"`
	IndexPatterns  []string `json:"index_patterns"`
	MaskedFields   []string `json:"masked_fields"`
}

// Role represents an Opensearch Role.
type Role struct {
	Hidden   bool `json:"hidden"`
	Reserved bool `json:"reserved"`
	Static   bool `json:"static"`
	RolePermissions
}

// RolePermissions contain only the permissions and description of the role.
// This subtype, which is embedded in Role, exists so that a valid PUT request
// can be easily made to the Opensearch API. This requires omitting the Hidden,
// Reserved, and Static fields.
type RolePermissions struct {
	ClusterPermissions []string           `json:"cluster_permissions"`
	Description        string             `json:"description,omitempty"`
	IndexPermissions   []IndexPermission  `json:"index_permissions"`
	TenantPermissions  []TenantPermission `json:"tenant_permissions"`
}

// RawRoles returns the raw JSON roles representation from the Opensearch API.
func (c *Client) RawRoles(ctx context.Context) ([]byte, error) {
	rolesURL := *c.baseURL
	rolesURL.Path = path.Join(c.baseURL.Path, "/_plugins/_security/api/roles/")
	req, err := http.NewRequestWithContext(ctx, "GET", rolesURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't construct roles request: %v", err)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't get roles: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bad roles response: %d\n%s", res.StatusCode, body)
	}
	return io.ReadAll(res.Body)
}

// Roles returns all Opensearch Roles.
func (c *Client) Roles(ctx context.Context) (map[string]Role, error) {
	data, err := c.RawRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't get roles from Opensearch API: %v", err)
	}
	var roles map[string]Role
	return roles, json.Unmarshal(data, &roles)
}

// CreateRole creates the given role in Opensearch.
func (c *Client) CreateRole(ctx context.Context, name string,
	role *Role) error {
	// Marshal payload. Payload only consists of RolePermissions because the
	// visibility fields are not writable.
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(role.RolePermissions); err != nil {
		return fmt.Errorf("couldn't marshal role: %v", err)
	}
	// construct request
	url := *c.baseURL
	url.Path = path.Join(c.baseURL.Path, "/_plugins/_security/api/roles/", name)
	req, err := http.NewRequestWithContext(ctx, "PUT", url.String(), &buf)
	if err != nil {
		return fmt.Errorf("couldn't construct create role request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// make request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't create role: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bad create role response: %d\n%s", res.StatusCode, body)
	}
	return nil
}

// DeleteRole deletes the named role from Opensearch.
func (c *Client) DeleteRole(ctx context.Context, name string) error {
	// construct request
	url := *c.baseURL
	url.Path = path.Join(c.baseURL.Path, "/_plugins/_security/api/roles/", name)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url.String(), nil)
	if err != nil {
		return fmt.Errorf("couldn't construct delete role request: %v", err)
	}
	// make request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't delete role: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bad delete role response: %d\n%s", res.StatusCode, body)
	}
	return nil
}
