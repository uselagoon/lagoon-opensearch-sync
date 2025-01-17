package keycloak

import "net/http"

// UseDefaultHTTPClient uses the default http client to avoid token refresh in
// tests.
func (c *Client) UseDefaultHTTPClient() {
	c.httpClient = http.DefaultClient
}
