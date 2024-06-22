package common

import (
    "runtime/debug"
    "vanity/proxy"
)

import (
    "context"
    "fmt"
    "github.com/sirupsen/logrus"
    stdio "net"
    "vanity/net"
)

type TcpListener struct {
    tag      string
    options  proxy.ListenerOptions
    listener *stdio.Listener
    defaults net.TcpOptions
}

func NewTcpListener(tag string, options net.TcpOptions) *TcpListener {
    return &TcpListener{
        tag:      tag,
        defaults: options,
    }
}

func (t *TcpListener) Tag() string {
    return t.tag
}

func (t *TcpListener) Network() net.Network {
    return net.Network_TCP
}

func (t *TcpListener) Init(options proxy.ListenerOptions) error {
    if options.Network != net.Network_TCP {
        return fmt.Errorf("% only support tcp network, was: %s", t.Tag(), options.Network)
    }
    t.options = options
    return nil
}

func (t *TcpListener) Serve(ctx context.Context, callback func(ctx context.Context, conn net.Connection)) error {
    addr := fmt.Sprintf("%s:%d", t.options.Address, t.options.Port)
    logrus.Infof("%s serve, address: %s", t.Tag(), addr)
    listener, err := stdio.Listen("tcp", addr)
    if err != nil {
        return fmt.Errorf("failed to listen tcp address %s %w", addr, err)
    }
    t.listener = &listener
    defer func() {
        logrus.Infof("%s terminated, address: %s", t.Tag(), addr)
        _ = listener.Close()
    }()
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            conn, err := listener.Accept()
            if err != nil {
                logrus.Errorf("%s accept error: %s", t.Tag(), err)
                return fmt.Errorf("%s accept: %w", t.Tag(), err)
            }
            go func(tcpConn *stdio.TCPConn) {
                defer func() {
                    if err := recover(); err != nil {
                        logrus.Errorf("%s handler err: %s, trace: %s", t.Tag(), err, string(debug.Stack()))
                    }
                }()
                defer func() {
                    logrus.Infof("%s connection terminated, address: %s", t.Tag(), tcpConn.RemoteAddr())
                    net.Close(tcpConn)
                }()
                if err := net.SetTcpOptions(tcpConn, t.defaults); err != nil {
                    logrus.Errorf("%s handler set local option: %s, trace: %s", t.Tag(), err, string(debug.Stack()))
                } else {
                    callback(ctx, net.Connection{
                        Address:         net.IPAddress((conn.RemoteAddr().(*stdio.TCPAddr)).IP),
                        TCPConn:         tcpConn,
                        ReadWriteCloser: conn,
                    })
                }
            }(conn.(*stdio.TCPConn))
        }
    }
}
