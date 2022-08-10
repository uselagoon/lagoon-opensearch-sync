package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
)

// Group represents a Keycloak Group. It holds the fields required when getting
// a list of groups from keycloak.
type Group struct {
	ID string `json:"id"`
	GroupUpdateRepresentation
}

// GroupUpdateRepresentation holds the fields required when updating a group.
type GroupUpdateRepresentation struct {
	Name       string              `json:"name"`
	Attributes map[string][]string `json:"attributes"`
}

// rawGroups returns the raw JSON group representation from the Keycloak API.
func (c *Client) rawGroups(ctx context.Context) ([]byte, error) {
	groupsURL := *c.baseURL
	groupsURL.Path = path.Join(c.baseURL.Path,
		"/auth/admin/realms/lagoon/groups")
	req, err := http.NewRequestWithContext(ctx, "GET", groupsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't construct groups request: %v", err)
	}
	q := req.URL.Query()
	q.Add("briefRepresentation", "false")
	req.URL.RawQuery = q.Encode()
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

// Groups returns all Keycloak Groups including their attributes.
func (c *Client) Groups(ctx context.Context) ([]Group, error) {
	rawGroups, err := c.rawGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't get groups from Keycloak API: %v", err)
	}
	var groups []Group
	return groups, json.Unmarshal(rawGroups, &groups)
}
