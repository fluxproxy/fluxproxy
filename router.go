package avoidy

import (
	"avoidy/net"
	"context"
)

type Router interface {
	Networks() []net.Network
	Route(ctx context.Context, conn net.Connection) error
}
