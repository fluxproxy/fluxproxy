package internal

import (
	"context"
	"fmt"
	"github.com/bytepowered/assert-go"
	"github.com/sirupsen/logrus"
	stdnet "net"
	"vanity/common"
	"vanity/net"
)

func TcpConnect(inctx context.Context, opts net.TcpOptions, target *net.Connection) error {
	assert.MustTrue(target.Destination.Network == net.Network_TCP, "unsupported network: %s", target.Destination.Network)
	logrus.Info("tcp-connector dail: ", target.Destination)
	localTcpConn := target.TCPConn
	remoteTCPConn, err := stdnet.DialTCP("tcp", nil, &stdnet.TCPAddr{
		IP:   target.Destination.Address.IP(),
		Port: int(target.Destination.Port),
	})
	if err != nil {
		return fmt.Errorf("tcp-connector dail: %w", err)
	}
	defer func() {
		logrus.Infof("tcp-connector dail-serve terminated, address: %s, %s ", target.Address, target.Destination)
		net.Close(remoteTCPConn)
	}()
	if err := net.SetTcpOptions(remoteTCPConn, opts); err != nil {
		return fmt.Errorf("tcp-connector set remote options: %w", err)
	}
	ctx, cancel := context.WithCancel(inctx)
	taskErrors := make(chan error, 2)
	send := func(_ context.Context, from, to net.Conn) {
		defer logrus.Info("tcp-connector send-loop terminated, destination: ", target.Destination)
		common.Copy(from, to, taskErrors)
	}
	receive := func(_ context.Context, from, to net.Conn) {
		defer logrus.Info("tcp-connector receive-loop terminated, destination: ", target.Destination)
		common.Copy(from, to, taskErrors)
	}
	go send(ctx, localTcpConn, remoteTCPConn)
	go receive(ctx, remoteTCPConn, localTcpConn)
	defer cancel()
	return <-taskErrors
}
