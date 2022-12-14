package opensearch

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"time"
)

// AuthenticatedRoundTripper implements the http.RoundTripper interface
type AuthenticatedRoundTripper struct {
	roundTripper http.RoundTripper
	username     string
	password     string
}

// RoundTrip sets the basic authentication header and then handles the request
// using a custom transport with which validates the connection using the
// configured CA.
func (art *AuthenticatedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(art.username, art.password)
	return art.roundTripper.RoundTrip(req)
}

// ca is the PEM encoded CA certificate
func httpClient(username, password string, ca *x509.Certificate) *http.Client {
	cp := x509.NewCertPool()
	cp.AddCert(ca)
	// construct http.Client with custom CA and automatic basic auth
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &AuthenticatedRoundTripper{
			roundTripper: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: cp,
				},
			},
			username: username,
			password: password,
		},
	}
}
