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
	go func() {
		defer listener.Close()
		defer func() {
			if rerr := recover(); rerr != nil {
				log.Printf("TcpListener crashed err: %s, \ntrace:%s", rerr, string(debug.Stack()))
			}
		}()
		for {
			conn, aerr := listener.Accept()
			if aerr != nil {
				log.Printf("TcpListener accept err: %s, \ntrace:%s", aerr, string(debug.Stack()))
				return
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
					Context:    connCtx,
					ReadWriter: conn,
					Address:    conn.RemoteAddr(),
					Network:    net.Network_TCP,
				})
			}()
		}
	}()
	return nil
}
