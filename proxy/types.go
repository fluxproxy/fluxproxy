package proxy

import (
	"context"
	"fluxway/net"
)

//// Server Type

type ServerType uint8

const (
	ServerType_RAWTCP ServerType = iota
	ServerType_RAWUDP
	ServerType_SOCKS5
	ServerType_HTTPS
)

//// Hook func

type HookFunc func(ctx context.Context, conn *net.Connection) error
