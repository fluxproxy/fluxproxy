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
	logger := proxy.LoggerFromContext(srcConnCtx)
	logger.Info("tcp-dail: ", link.Destination)
	srcTcpConn := link.TCPConn
	// 需要区分 TCPAddr / DomainAddr
	dstTCPConn, err := stdnet.DialTCP("tcp", nil, &stdnet.TCPAddr{
		IP:   link.Destination.Address.IP(),
		Port: int(link.Destination.Port),
	})
	if err != nil {
		return fmt.Errorf("tcp-dail: %w", err)
	}
	defer func() {
		logger.Infof("tcp-dail: serve terminated, address: %s, destination: %s", link.Address, link.Destination)
		net.Close(dstTCPConn)
	}()
	if err := net.SetTcpOptions(dstTCPConn, opts); err != nil {
		return fmt.Errorf("tcp-dail: set options: %w", err)
	}
	dstCtx, dstCancel := context.WithCancel(srcConnCtx)
	defer dstCancel()
	errors := make(chan error, 2)
	copier := func(_ context.Context, name string, from, to net.Conn) {
		defer logger.Info("tcp-dail: copier(%s) terminated, destination: ", name, link.Destination)
		errors <- net.Copier(from, to)
	}
	go copier(dstCtx, "src-to-dest", srcTcpConn, dstTCPConn)
	go copier(dstCtx, "dest-to-src", dstTCPConn, srcTcpConn)
	return <-errors
}
