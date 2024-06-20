package avoidy

import (
	"avoidy/net"
	"context"
)

type ListenerOptions struct {
	Network []net.Network
	Address string
	Port    int
}

type Listener interface {
	Init(options ListenerOptions) error
	Serve(ctx context.Context, handler func(ctx context.Context, conn net.Connection)) error
}
