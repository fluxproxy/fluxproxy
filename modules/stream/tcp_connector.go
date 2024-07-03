package stream

import (
	"context"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/helper"
	"github.com/rocketmanapp/rocket-proxy/net"
	"time"
)

var (
	_ rocket.Connector = (*TcpConnector)(nil)
)

type TcpConnector struct {
	opts net.TcpOptions
}

func NewTcpConnector() *TcpConnector {
	return &TcpConnector{
		opts: net.DefaultTcpOptions(),
	}
}

func (c *TcpConnector) DialServe(srcConnCtx context.Context, link *net.Connection) error {
	assert.MustTrue(link.Destination.Network == net.NetworkTCP, "dest network is not tcp, was: %s", link.Destination.Network)
	assert.MustTrue(link.Destination.Address.Family().IsIP(), "dest addr is not an ip, was: %s", link.Destination.Address.String())
	srcConn := link.TCPConn()
	dialer := &net.Dialer{
		Timeout:   time.Second * 5,
		KeepAlive: time.Duration(0),
	}
	dstConn, dErr := dialer.DialContext(srcConnCtx, "tcp", link.Destination.NetAddr())
	if dErr != nil {
		return fmt.Errorf("tcp-dial. %w", dErr)
	}
	defer helper.Close(dstConn)
	// Hook: dail
	if hook := rocket.LookupHookFunc(srcConnCtx, rocket.CtxHookFuncOnDialer); hook != nil {
		if hkErr := hook(srcConnCtx, link); hkErr != nil {
			return fmt.Errorf("tcp-dail:hook. %w", hkErr)
		}
	}
	ioErrors := make(chan error, 2)
	copier := func(name string, from, to net.Conn) {
		ioErrors <- helper.Copier(from, to)
	}
	go copier("src-to-dest", srcConn, dstConn)
	go copier("dest-to-src", dstConn, srcConn)
	return <-ioErrors
}
