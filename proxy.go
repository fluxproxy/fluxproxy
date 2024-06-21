package vanity

import (
	"context"
	"vanity/net"
)

type Inbound interface {
}

type Outbound interface {
}

type Router interface {
	Networks() []net.Network
	Route(ctx context.Context, conn net.Connection) error
}
