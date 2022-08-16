package opensearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
)

// RoleMapping represents an Opensearch RoleMapping.
type RoleMapping struct {
	AndBackendRoles []string `json:"and_backend_roles"`
	BackendRoles    []string `json:"backend_roles"`
	Hidden          bool     `json:"hidden"`
	Hosts           []string `json:"hosts"`
	Reserved        bool     `json:"reserved"`
	Users           []string `json:"users"`
}

// rawRolesmapping returns the raw JSON rolesmapping representation from the
// Opensearch API.
func (c *Client) rawRolesMapping(ctx context.Context) ([]byte, error) {
	rolesURL := *c.baseURL
	rolesURL.Path = path.Join(c.baseURL.Path,
		"/_plugins/_security/api/rolesmapping/")
	req, err := http.NewRequestWithContext(ctx, "GET", rolesURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't construct rolesmapping request: %v", err)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't get rolesmapping: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("bad rolesmapping response: %d\n%s",
			res.StatusCode, body)
	}
	return io.ReadAll(res.Body)
}

// RolesMapping returns all Opensearch RolesMapping.
func (c *Client) RolesMapping(
	ctx context.Context) (map[string]RoleMapping, error) {
	rawRolesMapping, err := c.rawRolesMapping(ctx)
	if err != nil {
		return nil,
			fmt.Errorf("couldn't get rolesmapping from Opensearch API: %v", err)
	}
	var rm map[string]RoleMapping
	return rm, json.Unmarshal(rawRolesMapping, &rm)
}
