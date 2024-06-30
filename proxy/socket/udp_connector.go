package socket

import (
	"context"
	"github.com/bytepowered/assert-go"
	"rocket/net"
	"rocket/proxy"
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
	assert.MustTrue(link.Destination.Network == net.Network_UDP, "dest network is not udp, was: %s", link.Destination.Network)
	// Hook: dail
	if hook := proxy.HookFuncDialPhased(srcConnCtx); hook != nil {
		if err := hook(srcConnCtx, link); err != nil {
			return err
		}
	}
	return nil
}
