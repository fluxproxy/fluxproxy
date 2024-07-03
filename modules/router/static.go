package router

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy/net"
)

//// Static target router

type StaticRouter struct {
	target net.Destination
}

func NewStaticRouter(target net.Destination) *StaticRouter {
	return &StaticRouter{
		target: target,
	}
}

func (d *StaticRouter) Route(ctx context.Context, income *net.Connection) (_ctx context.Context, target net.Connection, err error) {
	return ctx, income.WithDestination(d.target), nil
}
