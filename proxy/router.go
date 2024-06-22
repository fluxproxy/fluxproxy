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

func (d *StaticRouter) Router(ctx context.Context, income *net.Connection) (target net.Connection, err error) {
	dest := net.TCPDestination(net.LocalHostIP, net.Port(1234))
	return income.WithDestination(dest), nil
	//return *income, nil
}
