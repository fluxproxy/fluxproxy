package internal

import (
	"errors"
	"github.com/rocketmanapp/rocket-proxy/helper"
	"github.com/rocketmanapp/rocket-proxy/proxy"
	"runtime/debug"
	"time"
)

import (
	"context"
	"fmt"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	stdnet "net"
)

type TcpListener struct {
	tag     string
	options proxy.ListenerOptions
	tcpOpts net.TcpOptions
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
	listener, lErr := stdnet.ListenTCP("tcp", addr)
	if lErr != nil {
		return fmt.Errorf("failed to listen tcp address %s %w", addr, lErr)
	}
	go func() {
		<-serveCtx.Done()
		_ = listener.Close()
	}()
	var tempDelay time.Duration
	for {
		conn, aErr := listener.Accept()
		if aErr != nil {
			select {
			case <-serveCtx.Done():
				return serveCtx.Err()
			default:
				var netErr net.Error
				if errors.As(aErr, &netErr) && netErr.Temporary() {
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if maxDuration := 1 * time.Second; tempDelay > maxDuration {
						tempDelay = maxDuration
					}
					logrus.Errorf("http: Accept error: %v; retrying in %v", aErr, tempDelay)
					time.Sleep(tempDelay)
					continue
				}
				return fmt.Errorf("%s listen accept: %w", t.tag, aErr)
			}
		}
		go t.handle(serveCtx, conn.(*stdnet.TCPConn), handler)
	}
}

func (t *TcpListener) handle(serveCtx context.Context, tcpConn *stdnet.TCPConn, handler proxy.ListenerHandler) {
	defer func() {
		if rErr := recover(); rErr != nil {
			logrus.Errorf("%s handle conn: %s, trace: %s", t.tag, rErr, string(debug.Stack()))
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
	hErr := handler(connCtx, net.Connection{
		Network:     t.Network(),
		Address:     net.IPAddress((tcpConn.RemoteAddr().(*stdnet.TCPAddr)).IP),
		ReadWriter:  tcpConn,
		UserContext: context.Background(),
		Destination: net.DestinationNotset,
	})
	if hErr != nil {
		proxy.Logger(connCtx).Errorf("%s conn error: %s", t.tag, hErr)
	}
}
