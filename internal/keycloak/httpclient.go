package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func httpClient(ctx context.Context, u url.URL, realm, clientID,
	clientSecret string, timeout time.Duration) (*http.Client, error) {
	u.Path = path.Join(u.Path, fmt.Sprintf("/auth/realms/%s", realm))

	// Create a timeout-enabled HTTP client first
	timeoutClient := &http.Client{
		Timeout: timeout,
	}

	// Attach it to the context BEFORE discovery
	ctx = context.WithValue(ctx, oauth2.HTTPClient, timeoutClient)

	// Now discovery uses the timeout-enabled client
	provider, err := oidc.NewProvider(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("couldn't get new OIDC provider: %v", err)
	}

	c := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     provider.Endpoint().TokenURL,
	}

	client := c.Client(ctx)
	client.Timeout = timeout
	return client, nil
}
