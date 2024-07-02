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

//// Hook func

type HookFunc func(ctx context.Context, conn *net.Connection) error

////

var (
	_ ListenerHandler = (*ListenerHandlerAdapter)(nil)
)

type ListenerHandlerAdapter struct {
	Authorizer ListenerAuthorizeFunc
	Handler    ListenerHandlerFunc
}

func (l *ListenerHandlerAdapter) Handle(ctx context.Context, conn net.Connection) error {
	return l.Handler(ctx, conn)
}

func (l *ListenerHandlerAdapter) Authorize(ctx context.Context, conn net.Connection, auth ListenerAuthorization) error {
	return l.Authorizer(ctx, conn, auth)
}
