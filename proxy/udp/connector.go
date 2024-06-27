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

func (d *Connector) DialServe(srcConnCtx context.Context, link *net.Connection) (err error) {
	assert.MustTrue(link.Destination.Network == net.Network_UDP, "unsupported network: %s", link.Destination.Network)
	// Hook: dail
	if hook := proxy.LookupHookDialPhased(srcConnCtx); hook != nil {
		if err := hook(srcConnCtx, link); err != nil {
			return err
		}
	}
	return nil
}
