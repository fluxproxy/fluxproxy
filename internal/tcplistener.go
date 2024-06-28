package internal

import (
	"fluxway/helper"
	"fluxway/proxy"
	"runtime/debug"
)

import (
	"context"
	"fluxway/net"
	"fmt"
	"github.com/sirupsen/logrus"
	stdnet "net"
)

type TcpListener struct {
	tag      string
	options  proxy.ListenerOptions
	listener *stdnet.TCPListener
	tcpOpts  net.TcpOptions
}

func NewTcpListener(tag string, tcpOpts net.TcpOptions) *TcpListener {
	return &TcpListener{
		tag:     tag,
		tcpOpts: tcpOpts,
	}
}

func (t *TcpListener) ServerType() proxy.ServerType {
	return proxy.ServerType_RAWTCP
}

func (t *TcpListener) Network() net.Network {
	return net.Network_TCP
}

func (t *TcpListener) Init(options proxy.ListenerOptions) error {
	t.options = options
	return nil
}

func (t *TcpListener) Listen(serveCtx context.Context, handler proxy.ListenerHandler) error {
	addr := &stdnet.TCPAddr{IP: stdnet.ParseIP(t.options.Address), Port: t.options.Port}
	logrus.Infof("%s: listen start, address: %s", t.tag, addr)
	listener, err := stdnet.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen tcp address %s %w", addr, err)
	}
	t.listener = listener
	defer func() {
		logrus.Infof("%s listen stop, address: %s", t.tag, addr)
		_ = listener.Close()
	}()
	for {
		select {
		case <-serveCtx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				return fmt.Errorf("%s listen accept: %w", t.tag, err)
			}
			go func(tcpConn *stdnet.TCPConn) {
				defer func() {
					if err := recover(); err != nil {
						logrus.Errorf("%s handle conn: %s, trace: %s", t.tag, err, string(debug.Stack()))
					}
				}()
				// Set tcp conn options
				defer helper.Close(tcpConn)
				if err := net.SetTcpOptions(tcpConn, t.tcpOpts); err != nil {
					logrus.Errorf("%s set conn options: %s", t.tag, err)
					return
				}
				// Next
				connCtx, connCancel := context.WithCancel(serveCtx)
				defer connCancel()
				connCtx = SetupTcpContextLogger(serveCtx, tcpConn)
				err := handler(connCtx, net.Connection{
					Network:     t.Network(),
					Address:     net.IPAddress((conn.RemoteAddr().(*stdnet.TCPAddr)).IP),
					ReadWriter:  tcpConn,
					UserContext: context.Background(),
					Destination: net.DestinationNotset,
				})
				if err != nil {
					proxy.Logger(connCtx).Errorf("%s conn error: %s", t.tag, err)
				}
			}(conn.(*stdnet.TCPConn))
		}
	}
}
