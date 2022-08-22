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

// RoleMapping represents an Opensearch RoleMapping.
type RoleMapping struct {
	Hidden   bool `json:"hidden"`
	Reserved bool `json:"reserved"`
	RoleMappingPermissions
}

// RoleMappingPermissions contain only the permissions of the rolemapping.
// This subtype, which is embedded in RoleMapping, exists so that a valid PUT
// request can be easily made to the Opensearch API. This requires omitting the
// Hidden and Reserved fields.
type RoleMappingPermissions struct {
	AndBackendRoles []string `json:"and_backend_roles"`
	BackendRoles    []string `json:"backend_roles"`
	Hosts           []string `json:"hosts"`
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

// CreateRoleMapping creates the given rolemapping in Opensearch.
func (c *Client) CreateRoleMapping(ctx context.Context, name string,
	rm *RoleMapping) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	// Marshal payload. Payload only consists of RoleMappingPermissions because
	// the visibility fields are not writable.
	if err := enc.Encode(rm.RoleMappingPermissions); err != nil {
		return fmt.Errorf("couldn't marshal rolemapping: %v", err)
	}
	// construct request
	url := *c.baseURL
	url.Path = path.Join(c.baseURL.Path,
		"/_plugins/_security/api/rolesmapping/", name)
	req, err := http.NewRequestWithContext(ctx, "PUT", url.String(), &buf)
	if err != nil {
		return fmt.Errorf("couldn't construct create rolemapping request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// make request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't create rolemapping: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bad create rolemapping response: %d\n%s",
			res.StatusCode, body)
	}
	return nil
}

// DeleteRoleMapping deletes the named rolemapping from Opensearch.
func (c *Client) DeleteRoleMapping(ctx context.Context, name string) error {
	// construct request
	url := *c.baseURL
	url.Path = path.Join(c.baseURL.Path,
		"/_plugins/_security/api/rolesmapping/", name)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url.String(), nil)
	if err != nil {
		return fmt.Errorf("couldn't construct delete rolemapping request: %v", err)
	}
	// make request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("couldn't delete rolemapping: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode > 299 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("bad delete rolemapping response: %d\n%s", res.StatusCode, body)
	}
	return nil
}
