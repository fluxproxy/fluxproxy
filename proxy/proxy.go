package proxy

import (
	"context"
	"vanity/net"
)

type ListenerOptions struct {
	Network net.Network `json:"network"`
	Address string      `json:"address"`
	Port    int         `json:"port"`
}

type Listener interface {
	Network() net.Network
	Init(options ListenerOptions) error
	Serve(ctx context.Context, handler func(ctx context.Context, link net.Connection)) error
}

type Forwarder interface {
	DailServe(ctx context.Context, target *net.Link) (err error)
}

type Router interface {
	Router(ctx context.Context, conn *net.Connection) (target net.Link, err error)
}
