package authenticator

import (
	"context"
	"errors"
	"strings"
)

var (
	_ proxy.Authenticator = (*BasicAuthenticator)(nil)
)

var (
	ErrUPInvalidUsernameOrPassword = errors.New("basic:invalid username or password")
	ErrUPAuthenticateFailed        = errors.New("basic:authenticate failed")
)

type BasicAuthenticator struct {
	users map[string]string
}

func NewUsersAuthenticator(users map[string]string) *BasicAuthenticator {
	return &BasicAuthenticator{users: users}
}

func (u *BasicAuthenticator) Authenticate(ctx context.Context, auth proxy.Authentication) error {
	username, password, ok := strings.Cut(auth.Authentication, ":")
	if !ok {
		return ErrUPInvalidUsernameOrPassword
	}
	// check username and password
	if username == "" {
		return ErrUPInvalidUsernameOrPassword
	}
	if password == "" {
		return ErrUPInvalidUsernameOrPassword
	}
	if u.users[username] != password {
		return ErrUPAuthenticateFailed
	} else {
		return nil // success
	}
}
