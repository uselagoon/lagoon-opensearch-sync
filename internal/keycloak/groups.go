package keycloak

import (
	"context"
	"encoding/json"
	"errors"
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

// RawGroups returns the raw JSON group representation from the Keycloak API.
func (c *Client) RawGroups(ctx context.Context) ([]byte, error) {
	groupsURL := *c.baseURL
	groupsURL.Path = path.Join(c.baseURL.Path,
		"/auth/admin/realms/lagoon/groups")
	req, err := http.NewRequestWithContext(ctx, "GET", groupsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't construct groups request: %v", err)
	}
	q := req.URL.Query()
	q.Add("subGroupsCount", "false")
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
	data, err := c.RawGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't get groups from Keycloak API: %v", err)
	}
	var groups []Group
	if err = json.Unmarshal(data, &groups); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal groups from Keycloak API: %v", err)
	}
	if len(groups) == 0 {
		// https://github.com/uselagoon/lagoon-opensearch-sync/issues/150
		return nil,
			errors.New("empty groups response from Keycloak. Permissions issue?")
	}
	return groups, json.Unmarshal(data, &groups)
}
