package stream

import (
	"context"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/helper"
	"github.com/rocketmanapp/rocket-proxy/net"
	"io"
	stdnet "net"
)

var (
	_ rocket.Connector = (*Connector)(nil)
)

type Connector struct {
	opts net.UdpOptions
}

func NewUdpConnector() *Connector {
	return &Connector{
		opts: net.DefaultUdpOptions(),
	}
}

func (c *Connector) DialServe(srcConnCtx context.Context, link *net.Connection) (err error) {
	assert.MustTrue(link.Destination.Network == net.NetworkUDP, "dest network is not udp, was: %s", link.Destination.Network)
	srcRw := link.ReadWriter
	dstConn, err := stdnet.DialUDP("udp", nil, &stdnet.UDPAddr{IP: link.Destination.Address.IP(), Port: int(link.Destination.Port)})
	if err != nil {
		return fmt.Errorf("udp-dial: %w", err)
	}
	defer helper.Close(dstConn)
	if err := net.SetUdpConnOptions(dstConn, c.opts); err != nil {
		return fmt.Errorf("udp-dial: set options: %w", err)
	}
	dstCtx, dstCancel := context.WithCancel(srcConnCtx)
	defer dstCancel()
	// Hook: dail
	if hook := rocket.HookFuncDialPhased(srcConnCtx); hook != nil {
		if err := hook(srcConnCtx, link); err != nil {
			return err
		}
	}
	errors := make(chan error, 2)
	copier := func(_ context.Context, name string, from, to io.ReadWriter) {
		errors <- helper.Copier(from, to)
	}
	go copier(dstCtx, "src-to-dest", srcRw, dstConn)
	go copier(dstCtx, "dest-to-src", dstConn, srcRw)
	return <-errors
}
