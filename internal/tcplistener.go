package internal

import (
	"errors"
	"github.com/bytepowered/assert"
	"github.com/bytepowered/goes"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/helper"
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
	options rocket.ListenerOptions
	tcpOpts net.TcpOptions
}

func NewTcpListener(tag string, tcpOpts net.TcpOptions) *TcpListener {
	return &TcpListener{
		tag:     tag,
		tcpOpts: tcpOpts,
	}
}

func (t *TcpListener) Network() net.Network {
	return net.NetworkTCP
}

func (t *TcpListener) Init(options rocket.ListenerOptions) error {
	t.options = options
	return nil
}

func (t *TcpListener) Listen(serveCtx context.Context, dispatchHandler rocket.ListenerHandler) error {
	addr := &stdnet.TCPAddr{IP: stdnet.ParseIP(t.options.Address), Port: t.options.Port}
	logrus.Infof("%s: listen start, address: %s", t.tag, addr)
	listener, lErr := stdnet.ListenTCP("tcp", addr)
	if lErr != nil {
		return fmt.Errorf("failed to listen tcp address %s %w", addr, lErr)
	}
	_ = listener.SetDeadline(time.Time{})
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
					logrus.Errorf("https: Accept error: %v; retrying in %v", aErr, tempDelay)
					time.Sleep(tempDelay)
					continue
				}
				return fmt.Errorf("%s listen accept: %w", t.tag, aErr)
			}
		}
		goes.Go(func() {
			t.handle(serveCtx, conn.(*stdnet.TCPConn), dispatchHandler)
		})
	}
}

func (t *TcpListener) handle(serveCtx context.Context, tcpConn *stdnet.TCPConn, dispatchHandler rocket.ListenerHandler) {
	connCtx, connCancel := context.WithCancel(SetupTcpContextLogger(serveCtx, tcpConn))
	defer connCancel()
	defer func() {
		if rErr := recover(); rErr != nil {
			rocket.Logger(connCtx).Errorf("%s handle conn: %s, trace: %s", t.tag, rErr, string(debug.Stack()))
		}
	}()
	// Set tcp conn options
	defer helper.Close(tcpConn)
	if oErr := net.SetTcpConnOptions(tcpConn, t.tcpOpts); oErr != nil {
		rocket.Logger(connCtx).Errorf("%s set conn options: %s", t.tag, oErr)
		return
	}
	srcAddr := net.IPAddress((tcpConn.RemoteAddr().(*stdnet.TCPAddr)).IP)
	// Authenticate
	connCtx, aErr := dispatchHandler.Authenticate(connCtx, rocket.Authentication{
		Source:         srcAddr,
		Authenticate:   rocket.AuthenticateSource, // 源地址校验
		Authentication: tcpConn.RemoteAddr().String(),
	})
	if aErr != nil {
		rocket.Logger(connCtx).Errorf("%s auth error: %s", t.tag, aErr)
		return
	} else {
		assert.MustNotNil(connCtx, "authenticated context is nil")
	}
	// Next
	hErr := dispatchHandler.Dispatch(connCtx, net.Connection{
		Network:     t.Network(),
		Address:     srcAddr,
		ReadWriter:  tcpConn,
		UserContext: context.Background(),
		Destination: net.DestinationNotset,
	})
	if hErr != nil {
		rocket.Logger(connCtx).Errorf("%s conn error: %s", t.tag, hErr)
	}
}
