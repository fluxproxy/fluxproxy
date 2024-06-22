package udp

import (
	"context"
	"fluxway/net"
	"fluxway/proxy"
	"github.com/bytepowered/assert-go"
)

var (
	_ proxy.Connector = (*Connector)(nil)
)

type Connector struct {
}

func NewUdpConnector() *Connector {
	return &Connector{}
}

func (d *Connector) DailServe(ctx context.Context, target *net.Connection) (err error) {
	assert.MustTrue(target.Destination.Network == net.Network_UDP, "unsupported network: %s", target.Destination.Network)
	return nil
}
