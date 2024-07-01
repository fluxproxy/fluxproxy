package proxy

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy/net"
)

//// Server Type

type ServerType uint8

const (
	ServerType_RAWTCP ServerType = iota
	ServerType_RAWUDP
	ServerType_SOCKS5
	ServerType_HTTPS
)

func (t ServerType) String() string {
	switch t {
	case ServerType_RAWTCP:
		return "rawtcp"
	case ServerType_RAWUDP:
		return "rawudp"
	case ServerType_SOCKS5:
		return "socks5"
	case ServerType_HTTPS:
		return "https"
	}
	return "unknown"
}

//// Hook func

type HookFunc func(ctx context.Context, conn *net.Connection) error
