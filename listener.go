package vanity

import (
	"context"
	"vanity/net"
)

type ListenerOptions struct {
	Network net.Network
	Address string
	Port    int
}

type Listener interface {
	Tag() string
	Network() net.Network
	Init(options ListenerOptions) error
	Serve(ctx context.Context, handler func(ctx context.Context, link net.Connection)) error
}
