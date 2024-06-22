package socks5

import (
    "context"
    "github.com/things-go/go-socks5"
    "vanity/net"
    "vanity/proxy"
    "vanity/proxy/common"
)

var (
    _ proxy.Listener = (*Listener)(nil)
)

type Listener struct {
    *common.TcpListener
    server *socks5.Server
}

func NewListener() *Listener {
    return &Listener{
        TcpListener: common.NewTcpListener("socks-listener", net.DefaultTcpOptions()),
        server:      socks5.NewServer(),
    }
}

func (t *Listener) Serve(ctx context.Context, callback func(ctx context.Context, conn net.Connection)) error {
    return t.TcpListener.Serve(ctx, func(ctx context.Context, conn net.Connection) {
        //err := t.server.ServeConn(conn.TCPConn)
        //if err != nil {
        //    logrus.Errorf("socks-conn error: %s", err)
        //}

    })
}
