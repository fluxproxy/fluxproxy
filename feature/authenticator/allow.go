package authenticator

import (
	"context"
	"github.com/rocket-proxy/rocket-proxy"
)

var (
	_ rocket.Authenticator = (*AllowAuthenticator)(nil)
)

type AllowAuthenticator struct {
}

func NewAllowAuthenticator() *AllowAuthenticator {
	return &AllowAuthenticator{}
}

func (a *AllowAuthenticator) Authenticate(ctx context.Context, authentication rocket.Authentication) (context.Context, error) {
	return ctx, nil
}
