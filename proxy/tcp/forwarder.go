package tcp

import (
	"context"
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
}

func NewForwarder() *Forwarder {
	return &Forwarder{}
}

func (d *Forwarder) DailServe(ctx context.Context, target *net.Link) (err error) {
	assert.MustTrue(target.Destination.Network == net.Network_TCP, "unsupported network: %s", target.Destination.Network)
	remoteCtx, remoteCancel := context.WithCancel(ctx)
	defer remoteCancel()
	logrus.Info("tcp forwarder dail to server: ", target.Destination)
	remoteTCPConn, err := ionet.DialTCP("tcp", nil, &ionet.TCPAddr{
		IP:   target.Destination.Address.IP(),
		Port: int(target.Destination.Port),
	})
	if err != nil {
		return fmt.Errorf("tcp forwarder dail: %w", err)
	}
	localTCPConn := target.Connection.TCPConn
	defer remoteTCPConn.Close()
	send := func() error {
		_ = net.SetTcpConnOpts(localTCPConn, nil)
		defer logrus.Warn("tcp forwarder send loop stop, destination: ", target.Destination)
		return common.LoopTask(remoteCtx, func() error {
			_ = localTCPConn.SetReadDeadline(time.Now().Add(time.Second * 10))
			_ = remoteTCPConn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if _, err := io.Copy(remoteTCPConn, localTCPConn); err != nil {
				return fmt.Errorf("send to remote server: %w", err)
			}
			if err := net.SetTcpConnOpts(target.Connection.TCPConn, nil); err != nil {
				return fmt.Errorf("set tcp conn options error: %w", err)
			}
			return nil
		})
	}
	receive := func() error {
		_ = net.SetTcpConnOpts(remoteTCPConn, nil)
		defer logrus.Warn("tcp forwarder receive loop stop, destination: ", target.Destination)
		return common.LoopTask(remoteCtx, func() error {
			_ = remoteTCPConn.SetReadDeadline(time.Now().Add(time.Second * 10))
			_ = localTCPConn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if _, err := io.Copy(localTCPConn, remoteTCPConn); err != nil {
				return fmt.Errorf("receive from remote server: %w", err)
			}
			if err := net.SetTcpConnOpts(remoteTCPConn, nil); err != nil {
				return fmt.Errorf("set tcp conn options: %w", err)
			}
			return nil
		})
	}
	return common.RunTasks(remoteCtx, send, receive)
}
