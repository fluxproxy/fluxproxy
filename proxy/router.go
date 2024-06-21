package proxy

import (
	"context"
	"vanity/net"
)

////

var (
	_ Router = (*StaticRouter)(nil)
)

type StaticRouter struct {
}

func NewStaticRouter() *StaticRouter {
	return &StaticRouter{}
}

func (d *StaticRouter) Router(ctx context.Context, conn *net.Connection) (target net.Link, err error) {
	return net.Link{
		Connection:  conn,
		Destination: net.TCPDestination(net.LocalHostIP, net.Port(1234)),
	}, nil
}
