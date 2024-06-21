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
	defaults net.TcpOptions
}

func NewListener() *Listener {
	return &Listener{
		defaults: net.TcpOptions{
			ReadTimeout:  10,
			WriteTimeout: 10,
			ReadBuffer:   1024,
			WriteBuffer:  1024,
			NoDelay:      true,
			KeepAlive:    10,
		},
	}
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
				connCtx := ctx
				callback(connCtx, net.Connection{
					Address:         net.IPAddress((conn.RemoteAddr().(*ionet.TCPAddr)).IP),
					TCPConn:         conn.(*ionet.TCPConn),
					ReadWriteCloser: conn,
				})
			}()
		}
	}
}
