package auth

import (
	"context"
	"errors"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/net"
)

type UserAuthorizer struct {
	enabled bool
	users   map[string]string
}

func WithUsers(enabled bool, users map[string]string) *UserAuthorizer {
	return &UserAuthorizer{enabled: enabled, users: users}
}

func (u *UserAuthorizer) Authorize(ctx context.Context, conn net.Connection, auth rocket.ListenerAuthorization) error {
	if !u.enabled {
		return nil
	}
	// check username and password
	if auth.Username == "" {
		return errors.New("username is empty")
	}
	if auth.Password == "" {
		return errors.New("password is empty")
	}
	if u.users[auth.Username] != auth.Password {
		return errors.New("invalid username or password")
	}
	return nil
}
