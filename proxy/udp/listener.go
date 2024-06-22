package udp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
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
	listener *ionet.UDPConn
}

func NewListener() *Listener {
	return &Listener{}
}

func (t *Listener) Type() proxy.ProxyType {
	return proxy.ProxyType_RAWUDP
}

func (t *Listener) Network() net.Network {
	return net.Network_UDP
}

func (t *Listener) Init(options proxy.ListenerOptions) error {
	if options.Network != net.Network_UDP {
		return fmt.Errorf("udp-listener only support udp network")
	}
	t.options = options
	return nil
}

func (t *Listener) Serve(ctx context.Context, handler proxy.ListenerHandler) error {
	addr := &ionet.UDPAddr{IP: ionet.ParseIP(t.options.Address), Port: t.options.Port}
	logrus.Info("udp-listener serve: %s", addr)
	listener, err := ionet.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen udp address %s %w", addr, err)
	}
	defer func() {
		logrus.Info("udp-listener terminated, address: ", addr)
		_ = listener.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			var buffer = make([]byte, 2048)
			n, srcAddr, rerr := t.listener.ReadFromUDP(buffer)
			if rerr != nil {
				logrus.Info("udp-listener read err: %s", err)
				return fmt.Errorf("udp-listener error %w", rerr)
			}
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logrus.Errorf("udp-listener handler err: %s, trace: %s", err, string(debug.Stack()))
					}
				}()
				handler(ctx, net.Connection{
					Address:     net.IPAddress(srcAddr.IP),
					TCPConn:     nil,
					LongLive:    false,
					Destination: net.DestinationNotset,
					ReadWriteCloser: &wrapper{
						localAddr:  t.listener.LocalAddr(),
						remoteAddr: srcAddr,
						reader:     bytes.NewReader(buffer[:n]),
						writer: func(b []byte) (n int, err error) {
							return t.listener.WriteToUDP(b, srcAddr)
						},
					},
				})
			}()
		}
	}
}

var (
	_ ionet.Conn = (*wrapper)(nil)
)

type wrapper struct {
	localAddr  ionet.Addr
	remoteAddr ionet.Addr
	reader     io.Reader
	writer     func(b []byte) (n int, err error)
}

func (c *wrapper) Read(b []byte) (n int, err error) {
	return c.reader.Read(b)
}

func (c *wrapper) Write(b []byte) (n int, err error) {
	return c.writer(b)
}

func (c *wrapper) Close() error {
	return nil
}

func (c *wrapper) LocalAddr() ionet.Addr {
	return c.localAddr
}

func (c *wrapper) RemoteAddr() ionet.Addr {
	return c.remoteAddr
}

func (c *wrapper) SetDeadline(t time.Time) error {
	return nil
}

func (c *wrapper) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *wrapper) SetWriteDeadline(t time.Time) error {
	return nil
}
