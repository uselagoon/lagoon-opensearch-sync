package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"golang.org/x/oauth2"
)

// clientCredsTokenSource implements oauth2.TokenSource
type clientCredentialsTS struct {
	ClientID     string
	ClientSecret string
	Endpoint     oauth2.Endpoint
	Context      context.Context
}

func (c *clientCredentialsTS) Token() (*oauth2.Token, error) {
	config := &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     c.Endpoint,
	}
	ctx := context.WithValue(c.Context, oauth2.HTTPClient, &http.Client{
		Timeout: 30 * time.Second,
	})
	// Authenticate for a token. While technically it is possible to configure
	// keycloak to return a refresh token, according to the RFC this SHOULD NOT
	// be done. Instead if the token has expired we should re-auth using client
	// credentials.
	token, err := config.Exchange(ctx, "",
		oauth2.SetAuthURLParam("grant_type", "client_credentials"),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't get token for credentials: %v", err)
	}
	return token, nil
}

func httpClient(ctx context.Context, u url.URL, realm, clientID,
	clientSecret string) (*http.Client, error) {
	u.Path = path.Join(u.Path,
		fmt.Sprintf("/auth/realms/%s/protocol/openid-connect/token", realm))
	// construct tokensource
	ccts := &clientCredentialsTS{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: u.String(),
		},
		Context: ctx,
	}
	// Wrap tokensource in an auto-refresh cache layer, and then wrap the
	// tokensource in a *http.Client
	return oauth2.NewClient(ctx, oauth2.ReuseTokenSource(nil, ccts)), nil
}
