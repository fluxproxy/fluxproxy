package listener

import (
	"context"
	"fmt"
	"log"
	ionet "net"
	"runtime/debug"
	"vanity"
	"vanity/net"
)

var (
	_ vanity.Listener = (*TcpListener)(nil)
)

type TcpListener struct {
	options  vanity.ListenerOptions
	listener *ionet.Listener
}

func (t *TcpListener) Tag() string {
	return "tcp"
}

func (t *TcpListener) Network() net.Network {
	return t.options.Network
}

func (t *TcpListener) Init(options vanity.ListenerOptions) error {
	t.options = options
	return nil
}

func (t *TcpListener) Serve(ctx context.Context, callback func(ctx context.Context, conn net.Connection)) error {
	addr := fmt.Sprintf("%s:%d", t.options.Address, t.options.Port)
	log.Printf("TcpListener listen: %s", addr)
	listener, err := ionet.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen tcp address %s %w", addr, err)
	}
	t.listener = &listener
	defer func() {
		log.Printf("TcpListener terminaled: %s", addr)
		listener.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			conn, aerr := listener.Accept()
			if aerr != nil {
				log.Printf("TcpListener accept err: %s, \ntrace:%s", aerr, string(debug.Stack()))
				return fmt.Errorf("tcp listener error: %w", aerr)
			}
			go func() {
				defer func() {
					if cerr := recover(); cerr != nil {
						log.Printf("TcpListener handler err: %s, \ntrace:%s", cerr, string(debug.Stack()))
					}
				}()
				defer conn.Close()
				connCtx := ctx
				callback(connCtx, net.Connection{
					Context:         connCtx,
					Network:         t.Network(),
					Source:          conn.RemoteAddr(),
					Distinction:     nil,
					Conn:            conn.(*ionet.TCPConn),
					ReadWriteCloser: conn,
				})
			}()
		}
	}
}
