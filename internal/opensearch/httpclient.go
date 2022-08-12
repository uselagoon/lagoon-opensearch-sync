package opensearch

import (
	"net/http"
)

// AuthenticatedRoundTripper implements the http.RoundTripper interface
type AuthenticatedRoundTripper struct {
	roundTripper http.RoundTripper
	username     string
	password     string
}

// RoundTrip sets the basic authentication header and then handles the request
// using the http.DefaultTransport.
func (art *AuthenticatedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(art.username, art.password)
	return art.roundTripper.RoundTrip(req)
}

func httpClient(username, password string) *http.Client {
	// wrap the token in a *http.Client
	return &http.Client{
		Transport: &AuthenticatedRoundTripper{
			roundTripper: http.DefaultTransport,
			username:     username,
			password:     password,
		},
	}
}
