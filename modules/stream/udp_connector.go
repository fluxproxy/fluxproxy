package stream

import (
	"context"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/helper"
	"github.com/rocketmanapp/rocket-proxy/net"
	"io"
	"time"
)

var (
	_ rocket.Connector = (*UdpConnector)(nil)
)

type UdpConnector struct {
	opts net.UdpOptions
}

func NewUdpConnector() *UdpConnector {
	return &UdpConnector{
		opts: net.DefaultUdpOptions(),
	}
}

func (c *UdpConnector) DialServe(srcConnCtx context.Context, link *net.Connection) error {
	assert.MustTrue(link.Destination.Network == net.NetworkUDP, "dest network is not udp, was: %s", link.Destination.Network)
	srcRw := link.ReadWriter
	dialer := &net.Dialer{
		Timeout:   time.Second * 5,
		KeepAlive: time.Duration(0),
	}
	dstConn, dErr := dialer.DialContext(srcConnCtx, "udp", link.Destination.NetAddr())
	if dErr != nil {
		return fmt.Errorf("udp-dial. %w", dErr)
	}
	defer helper.Close(dstConn)
	// Hook: dail
	if hook := rocket.LookupHookFunc(srcConnCtx, rocket.CtxHookFuncOnDialer); hook != nil {
		if hkErr := hook(srcConnCtx, link); hkErr != nil {
			return fmt.Errorf("udp-dail:hook. %w", hkErr)
		}
	}
	ioErrors := make(chan error, 2)
	copier := func(name string, from, to io.ReadWriter) {
		ioErrors <- helper.Copier(from, to)
	}
	go copier("src-to-dest", srcRw, dstConn)
	go copier("dest-to-src", dstConn, srcRw)
	return <-ioErrors
}
