package authenticator

import (
	"context"
	"errors"
	"github.com/rocketmanapp/rocket-proxy"
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

func (u *BasicAuthenticator) Authenticate(ctx context.Context, auth rocket.Authentication) error {
	if !u.enabled {
		return nil
	}
	//assert.MustTrue(auth.Authenticate == rocket.AuthenticateBasic, "invalid auth type: %s", auth.Authenticate)
	username, password, ok := strings.Cut(auth.Authentication, ":")
	if !ok {
		return errors.New("invalid username or password")
	}
	// check username and password
	if username == "" {
		return errors.New("username is empty")
	}
	if password == "" {
		return errors.New("password is empty")
	}
	if u.users[username] != password {
		return errors.New("invalid username or password")
	}
	return nil
}
