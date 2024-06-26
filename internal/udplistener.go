package internal

import (
	"bytes"
	"context"
	"fluxway/net"
	"fluxway/proxy"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	ionet "net"
	"runtime/debug"
)

var (
	_ proxy.Listener = (*UdpListener)(nil)
)

type UdpListener struct {
	tag      string
	options  proxy.ListenerOptions
	listener *ionet.UDPConn
	udpOpts  net.UdpOptions
}

func NewUdpListener(tag string, udpOpts net.UdpOptions) *UdpListener {
	return &UdpListener{
		tag:     tag,
		udpOpts: udpOpts,
	}
}

func (t *UdpListener) ProxyType() proxy.ProxyType {
	return proxy.ProxyType_RAWUDP
}

func (t *UdpListener) Network() net.Network {
	return net.Network_UDP
}

func (t *UdpListener) Init(options proxy.ListenerOptions) error {
	t.options = options
	return nil
}

func (t *UdpListener) Serve(serveCtx context.Context, handler proxy.ListenerHandler) error {
	addr := &ionet.UDPAddr{IP: ionet.ParseIP(t.options.Address), Port: t.options.Port}
	logrus.Infof("udp-listener serve: %s", addr.String())
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
		case <-serveCtx.Done():
			return nil
		default:
			var buffer = make([]byte, 2048)
			n, srcAddr, rerr := t.listener.ReadFromUDP(buffer)
			if rerr != nil {
				logrus.Errorf("%s serve read: %s", t.tag, err)
				return fmt.Errorf("%s serve read: %w", t.tag, err)
			}
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logrus.Errorf("%s handle conn: %s, trace: %s", t.tag, err, string(debug.Stack()))
					}
				}()
				connCtx, connCancel := context.WithCancel(serveCtx)
				defer connCancel()
				connCtx = SetupUdpContextLogger(serveCtx, srcAddr)
				err := handler(connCtx, net.Connection{
					Network:     t.Network(),
					Address:     net.IPAddress(srcAddr.IP),
					TCPConn:     nil,
					Destination: net.DestinationNotset,
					ReadWriter: &wrapper{
						reader: bytes.NewReader(buffer[:n]),
						writer: func(b []byte) (n int, err error) {
							return t.listener.WriteToUDP(b, srcAddr)
						},
					},
				})
				if err != nil {
					logger := proxy.LoggerFromContext(connCtx)
					logger.Errorf("%s conn error: %s", t.tag, err)
				}
			}()
		}
	}
}

var (
	_ io.ReadWriter = (*wrapper)(nil)
)

type wrapper struct {
	reader io.Reader
	writer func(b []byte) (n int, err error)
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
