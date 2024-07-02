package auth

import (
	"context"
	"errors"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/net"
	"strings"
)

type BasicAuthorizer struct {
	enabled bool
	users   map[string]string
}

func WithBasicUsers(enabled bool, users map[string]string) *BasicAuthorizer {
	return &BasicAuthorizer{enabled: enabled, users: users}
}

func (u *BasicAuthorizer) Authorize(ctx context.Context, conn net.Connection, auth rocket.ListenerAuthorization) error {
	if !u.enabled {
		return nil
	}
	//assert.MustTrue(auth.Authenticate == rocket.AuthenticateBasic, "invalid auth type: %s", auth.Authenticate)
	username, password, ok := strings.Cut(auth.Authorization, ":")
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
