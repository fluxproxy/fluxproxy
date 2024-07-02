package proxy

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy/net"
)

//// Director Type

type ServerType uint8

const (
	ServerType_TCP ServerType = iota
	ServerType_UDP
	ServerType_SOCKS
	ServerType_HTTP
)

func (t ServerType) String() string {
	switch t {
	case ServerType_TCP:
		return "tcp"
	case ServerType_UDP:
		return "udp"
	case ServerType_SOCKS:
		return "socks"
	case ServerType_HTTP:
		return "https"
	}
	return "unknown"
}

//// Hook func

type HookFunc func(ctx context.Context, conn *net.Connection) error
