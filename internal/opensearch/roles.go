package opensearch

import (
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
	Name               string             `json:"-"` // ignore in marshaling
	ClusterPermissions []string           `json:"cluster_permissions"`
	Description        string             `json:"description,omitempty"`
	Hidden             bool               `json:"hidden"`
	IndexPermissions   []IndexPermission  `json:"index_permissions"`
	Reserved           bool               `json:"reserved"`
	Static             bool               `json:"static"`
	TenantPermissions  []TenantPermission `json:"tenant_permissions"`
}

// rawRoles returns the raw JSON roles representation from the Opensearch API.
func (c *Client) rawRoles(ctx context.Context) ([]byte, error) {
	rolesURL := *c.baseURL
	rolesURL.Path = path.Join(c.baseURL.Path, "/_plugins/_security/api/roles/")
	req, err := http.NewRequestWithContext(ctx, "GET", rolesURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't construct roles request: %v", err)
	}
	// q := req.URL.Query()
	// q.Add("briefRepresentation", "false")
	// req.URL.RawQuery = q.Encode()
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't get groups: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bad groups response: %d\n%s", res.StatusCode, body)
	}
	return io.ReadAll(res.Body)
}

// Roles returns all Opensearch Roles.
func (c *Client) Roles(ctx context.Context) (RoleSlice, error) {
	rawRoles, err := c.rawRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't get roles from Opensearch API: %v", err)
	}
	var roles RoleSlice
	return roles, json.Unmarshal(rawRoles, &roles)
}
