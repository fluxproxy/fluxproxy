package udp

import (
	"context"
	"github.com/bytepowered/assert-go"
	"vanity/net"
	"vanity/proxy"
)

var (
	_ proxy.Connector = (*Connector)(nil)
)

type Connector struct {
}

func NewConnector() *Connector {
	return &Connector{}
}

func (d *Connector) DailServe(ctx context.Context, target *net.Connection) (err error) {
	assert.MustTrue(target.Destination.Network == net.Network_UDP, "unsupported network: %s", target.Destination.Network)
	return nil
}
