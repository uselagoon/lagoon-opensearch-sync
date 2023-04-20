package keycloak

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2/clientcredentials"
)

func httpClient(ctx context.Context, u url.URL, realm, clientID,
	clientSecret string) (*http.Client, error) {
	u.Path = path.Join(u.Path, fmt.Sprintf("/auth/realms/%s", realm))
	provider, err := oidc.NewProvider(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("couldn't get new OIDC provider: %v", err)
	}
	c := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     provider.Endpoint().TokenURL,
	}
	return c.Client(ctx), nil
}
