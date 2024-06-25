package internal

import (
	"context"
	"fluxway/net"
	"fluxway/proxy"
	"fmt"
	"github.com/bytepowered/assert-go"
	stdnet "net"
)

func TcpDailServe(srcConnCtx context.Context, opts net.TcpOptions, link *net.Connection) error {
	assert.MustTrue(link.Destination.Network == net.Network_TCP, "unsupported network: %s", link.Destination.Network)
	assert.MustTrue(link.Destination.Address.Family().IsIP(), "destination address must be ip address, was: %s", link.Destination.Address.String())
	logger := proxy.LoggerFromContext(srcConnCtx)
	logger.Info("tcp-dail: ", link.Destination)
	srcConn := link.TCPConn
	dstConn, err := stdnet.Dial("tcp", link.Destination.NetAddr())
	if err != nil {
		return fmt.Errorf("tcp-dail: %w", err)
	}
	defer func() {
		logger.Debug("tcp-dail: dst conn closed")
		net.Close(dstConn)
	}()
	if err := net.SetTcpOptions(dstConn, opts); err != nil {
		return fmt.Errorf("tcp-dail: set options: %w", err)
	}
	dstCtx, dstCancel := context.WithCancel(srcConnCtx)
	defer dstCancel()
	errors := make(chan error, 2)
	copier := func(_ context.Context, name string, from, to net.Conn) {
		defer logger.Debugf("tcp-dail: %s copier completed", name)
		errors <- net.Copier(from, to)
	}
	go copier(dstCtx, "src-to-dest", srcConn, dstConn)
	go copier(dstCtx, "dest-to-src", dstConn, srcConn)
	return <-errors
}
