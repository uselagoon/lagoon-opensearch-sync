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

func httpClient(ctx context.Context, u url.URL, realm, clientID,
	clientSecret string) (*http.Client, error) {
	u.Path = path.Join(u.Path,
		fmt.Sprintf("/auth/realms/%s/protocol/openid-connect/token", realm))
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: u.String(),
		},
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Timeout: 10 * time.Second,
	})
	// authenticate for a token
	token, err := config.Exchange(ctx, "",
		oauth2.SetAuthURLParam("grant_type", "client_credentials"),
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't get token for credentials: %v", err)
	}
	// wrap the token in a *http.Client
	return config.Client(ctx, token), nil
}
