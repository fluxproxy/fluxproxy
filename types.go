package rocket

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy/net"
)

//// Director Type

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
