package tcp

import (
	"context"
	"fmt"
	"github.com/bytepowered/assert-go"
	"io"
	ionet "net"
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
	dailCtx, dailCtxCancel := context.WithCancel(ctx)
	defer dailCtxCancel()
	remoteTCPConn, err := ionet.DialTCP("tcp", nil, &ionet.TCPAddr{
		IP:   target.Destination.Address.IP(),
		Port: int(target.Destination.Port),
	})
	if err != nil {
		return fmt.Errorf("tcp dail error. %w", err)
	}
	send := func() error {
		return common.LoopTask(dailCtx, func() error {
			if err := net.SetTcpConnOpts(target.Connection.TCPConn, nil); err != nil {
				dailCtxCancel()
				return fmt.Errorf("set tcp conn options error. %w", err)
			}
			if _, err := io.Copy(remoteTCPConn, target.Connection.ReadWriteCloser); err != nil {
				dailCtxCancel()
				return fmt.Errorf("send to remote server error. %w", err)
			}
			return nil
		})
	}
	receive := func() error {
		return common.LoopTask(dailCtx, func() error {
			if err := net.SetTcpConnOpts(target.Connection.TCPConn, nil); err != nil {
				dailCtxCancel()
				return fmt.Errorf("set tcp conn options error. %w", err)
			}
			if _, err := io.Copy(target.Connection.ReadWriteCloser, remoteTCPConn); err != nil {
				dailCtxCancel()
				return fmt.Errorf("receive from remote server error. %w", err)
			}
			return nil
		})
	}
	return common.RunTasks(dailCtx, send, receive)
}
