package internal

import (
	"context"
	"fluxway/common"
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
	destTCPConn, err := stdnet.DialTCP("tcp", nil, &stdnet.TCPAddr{
		IP:   link.Destination.Address.IP(),
		Port: int(link.Destination.Port),
	})
	if err != nil {
		return fmt.Errorf("tcp-connector dail: %w", err)
	}
	defer func() {
		logrus.Infof("tcp-connector dail-serve terminated, address: %s, %s ", link.Address, link.Destination)
		net.Close(destTCPConn)
	}()
	if err := net.SetTcpOptions(destTCPConn, opts); err != nil {
		return fmt.Errorf("tcp-connector set remote options: %w", err)
	}
	destCtx, destCancel := context.WithCancel(srcConnCtx)
	taskErrors := make(chan error, 2)
	send := func(_ context.Context, from, to net.Conn) {
		defer logrus.Info("tcp-connector send-loop terminated, destination: ", link.Destination)
		common.Copy(from, to, taskErrors)
	}
	receive := func(_ context.Context, from, to net.Conn) {
		defer logrus.Info("tcp-connector receive-loop terminated, destination: ", link.Destination)
		common.Copy(from, to, taskErrors)
	}
	go send(destCtx, srcTcpConn, destTCPConn)
	go receive(destCtx, destTCPConn, srcTcpConn)
	defer destCancel()
	return <-taskErrors
}
