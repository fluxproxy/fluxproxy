package rocket

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy/net"
)

//// Server Type

type ServerType uint8

const (
	ServerTypeTCP ServerType = iota
	ServerTypeUDP
	ServerTypeSOCKS
	ServerTypeHTTPS
)

func (t ServerType) String() string {
	switch t {
	case ServerTypeTCP:
		return "tcp"
	case ServerTypeUDP:
		return "udp"
	case ServerTypeSOCKS:
		return "socks"
	case ServerTypeHTTPS:
		return "https"
	}
	return "unknown"
}

//// Authenticate Types

const (
	AuthenticateSource = "Source"
	AuthenticateBasic  = "Basic"
	AuthenticateBearer = "Bearer"
	AuthenticateToken  = "Token"
)

//// Hook func

type HookFunc func(ctx context.Context, conn *net.Connection) error

//// ListenerHandlerAdapter

var (
	_ ListenerHandler = (*ListenerHandlerAdapter)(nil)
)

type ListenerHandlerAdapter struct {
	Authenticator AuthenticateFunc
	Dispatcher    DispatchFunc
}

func (l *ListenerHandlerAdapter) Dispatch(ctx context.Context, conn net.Connection) error {
	return l.Dispatcher(ctx, conn)
}

func (l *ListenerHandlerAdapter) Authenticate(ctx context.Context, conn net.Connection, auth Authentication) error {
	return l.Authenticator(ctx, conn, auth)
}
