package tcp

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/assert-go"
	"github.com/sirupsen/logrus"
	"io"
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
			AwaitTimeout: time.Millisecond * 50,
		},
	}
}

func (d *Forwarder) DailServe(ctx context.Context, target *net.Link) (err error) {
	assert.MustTrue(target.Destination.Network == net.Network_TCP, "unsupported network: %s", target.Destination.Network)
	remoteCtx, remoteCancel := context.WithCancel(ctx)
	defer remoteCancel()
	logrus.Info("tcp-forwarder dail: ", target.Destination)
	remoteTCPConn, err := ionet.DialTCP("tcp", nil, &ionet.TCPAddr{
		IP:   target.Destination.Address.IP(),
		Port: int(target.Destination.Port),
	})
	if err != nil {
		return fmt.Errorf("tcp-forwarder dail: %w", err)
	}
	defer func() {
		logrus.Infof("tcp-forwarder dail-serve terminated, address: %s, %s ", target.Connection.Address, target.Destination)
		net.Close(remoteTCPConn)
	}()
	if err := net.SetTcpOptions(remoteTCPConn, d.defaults); err != nil {
		return fmt.Errorf("tcp-forwarder set remote options: %w", err)
	}
	localTCPConn := target.Connection.TCPConn
	send := func() error {
		defer logrus.Warn("tcp-forwarder send-loop terminated, destination: ", target.Destination)
		_ = localTCPConn.SetDeadline(time.Time{})
		defer remoteCancel()
		if _, err := io.Copy(remoteTCPConn, localTCPConn); err == nil {
			return nil // A successful copy end
		} else {
			return checkConnError("local-conn", err)
		}
	}
	receive := func() error {
		defer logrus.Warn("tcp-forwarder receive-loop terminated, destination: ", target.Destination)
		_ = remoteTCPConn.SetDeadline(time.Time{})
		defer remoteCancel()
		if _, err := io.Copy(localTCPConn, remoteTCPConn); err == nil {
			return nil // A successful copy end
		} else {
			return checkConnError("remote-conn", err)
		}
	}
	return common.RunTasks(remoteCtx, send, receive)
}

func checkConnError(tag string, err error) error {
	cause := common.ErrorCause(err)
	if cause == io.EOF {
		return fmt.Errorf("%s end", tag)
	}
	if errors.Is(cause, ionet.ErrClosed) {
		return fmt.Errorf("%s closed", tag)
	}
	if net.IsTimeoutErr(cause) {
		return nil
	} else {
		return fmt.Errorf("%s receive: %w", tag, err)
	}
}
