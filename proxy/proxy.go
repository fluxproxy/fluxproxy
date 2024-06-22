package proxy

import (
	"context"
	"vanity/net"
)

type ProxyType uint8

const (
	ProxyType_RAWTCP ProxyType = iota
	ProxyType_RAWUDP
	ProxyType_SOCKS5
	ProxyType_HTTPS
)

type ListenerOptions struct {
	Network net.Network `json:"network"`
	Address string      `json:"address"`
	Port    int         `json:"port"`
}

type ListenerHandler func(ctx context.Context, conn net.Connection)

type Listener interface {
	Network() net.Network
	Type() ProxyType
	Init(options ListenerOptions) error
	Serve(ctx context.Context, handler ListenerHandler) error
}

type Connector interface {
	DailServe(ctx context.Context, target *net.Connection) (err error)
}

type Dispatcher interface {
	Dispatch(ctx context.Context, income *net.Connection) (target net.Connection, err error)
}
