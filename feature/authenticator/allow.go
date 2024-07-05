package authenticator

import (
	"context"
	"errors"
)

var (
	_ proxy.Authenticator = (*AllowAuthenticator)(nil)
)

type AllowAuthenticator struct {
}

func NewAllowAuthenticator() *AllowAuthenticator {
	return &AllowAuthenticator{}
}

func (a *AllowAuthenticator) Authenticate(ctx context.Context, authentication proxy.Authentication) error {
	return nil
}

////

type DenyAuthenticator struct {
}

func NewDenyAuthenticator() *DenyAuthenticator {
	return &DenyAuthenticator{}
}

func (a *DenyAuthenticator) Authenticate(ctx context.Context, authentication proxy.Authentication) error {
	return errors.New("authenticate deny for all")
}
