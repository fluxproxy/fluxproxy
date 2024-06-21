package listener

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	ionet "net"
	"runtime/debug"
	"time"
	"vanity"
	"vanity/net"
)

var (
	_ vanity.Listener = (*UdpListener)(nil)
)

type UdpListener struct {
	options  vanity.ListenerOptions
	listener *ionet.UDPConn
}

func (t *UdpListener) Tag() string {
	return "udp"
}

func (t *UdpListener) Network() net.Network {
	return t.options.Network
}

func (t *UdpListener) Init(options vanity.ListenerOptions) error {
	t.options = options
	return nil
}

func (t *UdpListener) Serve(ctx context.Context, handler func(ctx context.Context, conn net.Connection)) error {
	addr := &ionet.UDPAddr{IP: ionet.ParseIP(t.options.Address), Port: t.options.Port}
	log.Printf("UdpListener listen: %s", addr)
	listener, err := ionet.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen udp address %s %w", addr, err)
	}
	defer listener.Close()
	defer func() {
		if rerr := recover(); rerr != nil {
			log.Printf("UdpListener crashed err: %s, \ntrace:%s", rerr, string(debug.Stack()))
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			var buffer = make([]byte, 2048)
			n, srcAddr, rerr := t.listener.ReadFromUDP(buffer)
			if rerr != nil {
				log.Printf("UdpListener read err: %s, \ntrace:%s", err, string(debug.Stack()))
				return fmt.Errorf("udp listener error %w", rerr)
			}
			connCtx := ctx
			go handler(connCtx, net.Connection{
				Context:     connCtx,
				Source:      srcAddr,
				Distinction: nil,
				Network:     t.Network(),
				Conn:        nil,
				ReadWriteCloser: &wrapper{
					localAddr:  t.listener.LocalAddr(),
					remoteAddr: srcAddr,
					reader:     bytes.NewReader(buffer[:n]),
					writer: func(b []byte) (n int, err error) {
						return t.listener.WriteToUDP(b, srcAddr)
					},
				},
			})
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

//// convert func to io.Writer

//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
//})
//func (c *udpconn) WriteTo(b []byte) (n int, err error) {
//	return c.writer.Write(b)
//}
