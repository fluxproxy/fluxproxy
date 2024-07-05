package authenticator

import (
	"context"
	"errors"
	"github.com/rocket-proxy/rocket-proxy"
	"strings"
)

var (
	_ rocket.Authenticator = (*BasicAuthenticator)(nil)
)

type BasicAuthenticator struct {
	enabled bool
	users   map[string]string
}

func WithBasicUsers(enabled bool, users map[string]string) *BasicAuthenticator {
	return &BasicAuthenticator{enabled: enabled, users: users}
}

func (u *BasicAuthenticator) Authenticate(ctx context.Context, auth rocket.Authentication) (context.Context, error) {
	if !u.enabled {
		return ctx, nil
	}
	//assert.MustTrue(auth.Authenticate == rocket.AuthenticateBasic, "invalid auth type: %s", auth.Authenticate)
	username, password, ok := strings.Cut(auth.Authentication, ":")
	if !ok {
		return ctx, errors.New("invalid username or password")
	}
	// check username and password
	if username == "" {
		return ctx, errors.New("username is empty")
	}
	if password == "" {
		return ctx, errors.New("password is empty")
	}
	if u.users[username] != password {
		return ctx, errors.New("invalid username or password")
	}
	return ctx, nil
}
