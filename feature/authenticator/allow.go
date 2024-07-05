package authenticator

import (
	"context"
	"errors"
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

func (a *AllowAuthenticator) Authenticate(ctx context.Context, authentication rocket.Authentication) error {
	return nil
}

////

type DenyAuthenticator struct {
}

func NewDenyAuthenticator() *DenyAuthenticator {
	return &DenyAuthenticator{}
}

func (a *DenyAuthenticator) Authenticate(ctx context.Context, authentication rocket.Authentication) error {
	return errors.New("authenticate deny for all")
}
