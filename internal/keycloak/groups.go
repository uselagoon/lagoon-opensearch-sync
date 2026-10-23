package keycloak

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
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

// RawGroups returns the raw JSON group representation from the Keycloak API
// for a single page of results.
//
// - first is the pagination offset (the index of the first result to return),
// - max is the maximum number of results to return in this page.
//
// https://www.keycloak.org/docs-api/latest/rest-api/index.html#_get_adminrealmsrealmgroups
func (c *Client) RawGroups(ctx context.Context, first, max uint) ([]byte, error) {
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
	q.Add("first", strconv.FormatUint(uint64(first), 10))
	q.Add("max", strconv.FormatUint(uint64(max), 10))
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
//
// Groups are fetched from the Keycloak Admin API in pages of up to
// c.groupsPageSize groups at a time, using the first/max query parameters
// to page through the full result set.
func (c *Client) Groups(ctx context.Context) ([]Group, error) {
	if c.groupsPageSize == 0 {
		return nil, errors.New("groupsPageSize must be greater than zero")
	}
	var groups []Group
	for first := uint(0); ; first += c.groupsPageSize {
		data, err := c.RawGroups(ctx, first, c.groupsPageSize)
		if err != nil {
			return nil, fmt.Errorf("couldn't get groups from Keycloak API: %v", err)
		}
		var page []Group
		if err = json.Unmarshal(data, &page); err != nil {
			return nil, fmt.Errorf("couldn't unmarshal groups from Keycloak API: %v", err)
		}
		groups = append(groups, page...)
		if uint(len(page)) < c.groupsPageSize {
			// short (or empty) page: this was the last page of results
			break
		}
	}
	if len(groups) == 0 {
		// https://github.com/uselagoon/lagoon-opensearch-sync/issues/150
		return nil,
			errors.New("empty groups response from Keycloak. Permissions issue?")
	}
	return groups, nil
}
