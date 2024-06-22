package tcp

import (
	"context"
	"time"
	"vanity/internal"
	"vanity/net"
	"vanity/proxy"
)

var (
	_ proxy.Connector = (*Connector)(nil)
)

type Connector struct {
	opts net.TcpOptions
}

func NewTcpConnector() *Connector {
	return &Connector{
		opts: net.TcpOptions{
			ReadTimeout:  time.Second * 30,
			WriteTimeout: time.Second * 10,
			ReadBuffer:   1024,
			WriteBuffer:  1024,
			NoDelay:      true,
			KeepAlive:    time.Second * 10,
		},
	}
}

func (d *Connector) DailServe(inctx context.Context, target *net.Connection) error {
	return internal.TcpConnect(inctx, d.opts, target)
}
