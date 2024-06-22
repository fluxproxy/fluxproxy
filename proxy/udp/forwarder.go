package udp

import (
	"context"
	"github.com/bytepowered/assert-go"
	"vanity/net"
	"vanity/proxy"
)

var (
	_ proxy.Forwarder = (*Forwarder)(nil)
)

type Forwarder struct {
}

func (d *Forwarder) DailServe(ctx context.Context, target *net.Connection) (err error) {
	assert.MustTrue(target.Destination.Network == net.Network_UDP, "unsupported network: %s", target.Destination.Network)
	return nil
}
