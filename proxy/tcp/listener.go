package tcp

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	ionet "net"
	"runtime/debug"
	"vanity/net"
	"vanity/proxy"
)

var (
	_ proxy.Listener = (*Listener)(nil)
)

type Listener struct {
	options  proxy.ListenerOptions
	listener *ionet.Listener
}

func NewListener() *Listener {
	return &Listener{}
}

func (t *Listener) Tag() string {
	return "tcp"
}

func (t *Listener) Network() net.Network {
	return net.Network_TCP
}

func (t *Listener) Init(options proxy.ListenerOptions) error {
	if options.Network != net.Network_TCP {
		return fmt.Errorf("tcp listener only support tcp network")
	}
	t.options = options
	return nil
}

func (t *Listener) Serve(ctx context.Context, callback func(ctx context.Context, conn net.Connection)) error {
	addr := fmt.Sprintf("%s:%d", t.options.Address, t.options.Port)
	logrus.Info("tcp listener serve, address: ", addr)
	listener, err := ionet.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen tcp address %s %w", addr, err)
	}
	t.listener = &listener
	defer func() {
		logrus.Info("tcp listener terminated, address: ", addr)
		listener.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				logrus.Error("tcp listener accept error:", err)
				return fmt.Errorf("tcp listener error: %w", err)
			}
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logrus.Errorf("tcp listener handler err: %s, trace: %s", err, string(debug.Stack()))
					}
				}()
				defer conn.Close()
				tcpConn := conn.(*ionet.TCPConn)
				if err := net.SetTcpConnOpts(tcpConn, nil); err != nil {
					logrus.Error("tcp listener set options error:", err)
					return
				}
				connCtx := ctx
				callback(connCtx, net.Connection{
					Address:         net.IPAddress((conn.RemoteAddr().(*ionet.TCPAddr)).IP),
					TCPConn:         tcpConn,
					ReadWriteCloser: conn,
				})
			}()
		}
	}
}
