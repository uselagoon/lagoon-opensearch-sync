package dashboards

import (
	"net/http"
	"time"
)

// AuthenticatedRoundTripper implements the http.RoundTripper interface
type AuthenticatedRoundTripper struct {
	username string
	password string
}

// RoundTrip sets the basic authentication header and then handles the request
// using the http.DefaultTransport.
func (art *AuthenticatedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(art.username, art.password)
	return http.DefaultTransport.RoundTrip(req)
}

func httpClient(username, password string) *http.Client {
	// construct http.Client with automatic basic auth
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &AuthenticatedRoundTripper{
			username: username,
			password: password,
		},
	}
}
