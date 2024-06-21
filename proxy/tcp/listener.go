package tcp

import (
	"context"
	"fmt"
	"log"
	ionet "net"
	"runtime/debug"
	"time"
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
	log.Printf("tcp listener listen: %s", addr)
	listener, err := ionet.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen tcp address %s %w", addr, err)
	}
	t.listener = &listener
	defer func() {
		log.Printf("tcp listener terminaled: %s", addr)
		listener.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, aerr := listener.Accept()
			if aerr != nil {
				log.Printf("tcp listener accept err: %s, \ntrace:%s", aerr, string(debug.Stack()))
				return fmt.Errorf("tcp listener error: %w", aerr)
			}
			go func() {
				defer func() {
					if cerr := recover(); cerr != nil {
						log.Printf("tcp listener handler err: %s, \ntrace:%s", cerr, string(debug.Stack()))
					}
				}()
				defer conn.Close()
				tcpConn := conn.(*ionet.TCPConn)
				_ = tcpConn.SetReadDeadline(time.Now().Add(time.Second * 10))
				_ = tcpConn.SetWriteDeadline(time.Now().Add(time.Second * 10))
				_ = tcpConn.SetKeepAlive(true)
				_ = tcpConn.SetNoDelay(true)
				_ = tcpConn.SetReadBuffer(1024)
				_ = tcpConn.SetWriteBuffer(1024)
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
