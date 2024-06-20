package listener

import (
	"avoidy"
	"avoidy/net"
	"context"
	"fmt"
	"log"
	ionet "net"
	"runtime/debug"
)

var (
	_ avoidy.Listener = (*TcpListener)(nil)
)

type TcpListener struct {
	options  avoidy.ListenerOptions
	listener *ionet.Listener
}

func (t *TcpListener) Tag() string {
	return "tcp"
}

func (t *TcpListener) Network() net.Network {
	return t.options.Network
}

func (t *TcpListener) Init(options avoidy.ListenerOptions) error {
	t.options = options
	return nil
}

func (t *TcpListener) Serve(ctx context.Context, handler func(ctx context.Context, conn net.Connection)) error {
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
	defer func() {
		if rerr := recover(); rerr != nil {
			log.Printf("TcpListener crashed err: %s, \ntrace:%s", rerr, string(debug.Stack()))
		}
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
				handler(connCtx, net.Connection{
					Context:         connCtx,
					ReadWriteCloser: conn,
					Source:          conn.RemoteAddr(),
					Network:         net.Network_TCP,
				})
			}()
		}
	}
}
