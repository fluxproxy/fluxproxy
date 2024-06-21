package vanity

import (
	"context"
	"vanity/net"
)

type Inbound interface {
	Process(ctx context.Context, conn *net.Connection) (err error)
}

type Outbound interface {
	DailServe(ctx context.Context, target *net.Link) (err error)
}

type Router interface {
	Router(ctx context.Context, conn *net.Connection) (target net.Link, err error)
}
