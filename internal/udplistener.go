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
	tag     string
	options proxy.ListenerOptions
	udpOpts net.UdpOptions
}

func NewUdpListener(tag string, udpOpts net.UdpOptions) *UdpListener {
	return &UdpListener{
		tag:     tag,
		udpOpts: udpOpts,
	}
}

func (t *UdpListener) ServerType() proxy.ServerType {
	return proxy.ServerType_RAWUDP
}

func (t *UdpListener) Network() net.Network {
	return net.Network_UDP
}

func (t *UdpListener) Init(options proxy.ListenerOptions) error {
	t.options = options
	return nil
}

func (t *UdpListener) Listen(serveCtx context.Context, next proxy.ListenerHandler) error {
	addr := &ionet.UDPAddr{IP: ionet.ParseIP(t.options.Address), Port: t.options.Port}
	logrus.Infof("%s: listen start, address: %s", t.tag, addr)
	listener, err := ionet.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen udp address %s %w", addr, err)
	}
	go func() {
		<-serveCtx.Done()
		_ = listener.Close()
	}()
	for {
		var buffer = make([]byte, 1024*32)
		n, srcAddr, rerr := listener.ReadFromUDP(buffer)
		if rerr != nil {
			select {
			case <-serveCtx.Done():
				return nil
			default:
			}
			return fmt.Errorf("%s listen read: %w", t.tag, err)
		}
		go func() {
			defer func() {
				if err := recover(); err != nil {
					logrus.Errorf("%s handle conn: %s, trace: %s", t.tag, err, string(debug.Stack()))
				}
			}()
			// Next
			connCtx, connCancel := context.WithCancel(serveCtx)
			defer connCancel()
			connCtx = SetupUdpContextLogger(serveCtx, srcAddr)
			err := next(connCtx, net.Connection{
				Network:     t.Network(),
				Address:     net.IPAddress(srcAddr.IP),
				UserContext: context.Background(),
				ReadWriter: &wrapper{
					reader: bytes.NewReader(buffer[:n]),
					writer: func(b []byte) (n int, err error) {
						return listener.WriteToUDP(b, srcAddr)
					},
				},
				Destination: net.DestinationNotset,
			})
			if err != nil {
				proxy.Logger(connCtx).Errorf("%s conn error: %s", t.tag, err)
			}
		}()
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
