package tcp

import (
	"context"
	"fmt"
	"github.com/bytepowered/assert-go"
	"github.com/sirupsen/logrus"
	ionet "net"
	"time"
	"vanity/common"
	"vanity/net"
	"vanity/proxy"
)

var (
	_ proxy.Forwarder = (*Forwarder)(nil)
)

type Forwarder struct {
	defaults net.TcpOptions
}

func NewForwarder() *Forwarder {
	return &Forwarder{
		defaults: net.TcpOptions{
			ReadTimeout:  time.Second * 30,
			WriteTimeout: time.Second * 10,
			ReadBuffer:   1024,
			WriteBuffer:  1024,
			NoDelay:      true,
			KeepAlive:    time.Second * 10,
		},
	}
}

func (d *Forwarder) DailServe(inctx context.Context, target *net.Connection) error {
	assert.MustTrue(target.Destination.Network == net.Network_TCP, "unsupported network: %s", target.Destination.Network)
	logrus.Info("tcp-forwarder dail: ", target.Destination)
	remoteTCPConn, err := ionet.DialTCP("tcp", nil, &ionet.TCPAddr{
		IP:   target.Destination.Address.IP(),
		Port: int(target.Destination.Port),
	})
	if err != nil {
		return fmt.Errorf("tcp-forwarder dail: %w", err)
	}
	defer func() {
		logrus.Infof("tcp-forwarder dail-serve terminated, address: %s, %s ", target.Address, target.Destination)
		net.Close(remoteTCPConn)
	}()
	if err := net.SetTcpOptions(remoteTCPConn, d.defaults); err != nil {
		return fmt.Errorf("tcp-forwarder set remote options: %w", err)
	}
	ctx, cancel := context.WithCancel(inctx)
	taskErrors := make(chan error, 2)
	send := func(_ context.Context) {
		defer logrus.Info("tcp-forwarder send-loop terminated, destination: ", target.Destination)
		common.Copy(target.TCPConn, remoteTCPConn, taskErrors)
	}
	receive := func(_ context.Context) {
		defer logrus.Info("tcp-forwarder receive-loop terminated, destination: ", target.Destination)
		common.Copy(remoteTCPConn, target.TCPConn, taskErrors)
	}
	go send(ctx)
	go receive(ctx)
	defer cancel()
	return <-taskErrors
}
