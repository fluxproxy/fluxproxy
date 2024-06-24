package internal

import (
	"context"
	"fluxway/net"
	"fmt"
	"github.com/bytepowered/assert-go"
	"github.com/sirupsen/logrus"
	stdnet "net"
)

func TcpConnect(srcConnCtx context.Context, opts net.TcpOptions, link *net.Connection) error {
	assert.MustTrue(link.Destination.Network == net.Network_TCP, "unsupported network: %s", link.Destination.Network)
	logrus.Info("tcp-connector dail: ", link.Destination)
	srcTcpConn := link.TCPConn
	dstTCPConn, err := stdnet.DialTCP("tcp", nil, &stdnet.TCPAddr{
		IP:   link.Destination.Address.IP(),
		Port: int(link.Destination.Port),
	})
	if err != nil {
		return fmt.Errorf("tcp-connector dail: %w", err)
	}
	defer func() {
		logrus.Infof("tcp-connector dail-serve terminated, address: %s, %s ", link.Address, link.Destination)
		net.Close(dstTCPConn)
	}()
	if err := net.SetTcpOptions(dstTCPConn, opts); err != nil {
		return fmt.Errorf("tcp-connector set remote options: %w", err)
	}
	dstCtx, dstCancel := context.WithCancel(srcConnCtx)
	defer dstCancel()
	errors := make(chan error, 2)
	copier := func(_ context.Context, name string, from, to net.Conn) {
		defer logrus.Info("tcp-connector %s copier terminated, destination: ", name, link.Destination)
		errors <- net.Copier(from, to)
	}
	go copier(dstCtx, "src-to-dest", srcTcpConn, dstTCPConn)
	go copier(dstCtx, "dest-to-src", dstTCPConn, srcTcpConn)
	return <-errors
}
