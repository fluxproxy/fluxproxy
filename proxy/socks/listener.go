package socks

import "C"
import (
	"context"
	"fluxway/internal"
	"fluxway/net"
	"fluxway/proxy"
	"fmt"
	"github.com/riobard/go-shadowsocks2/socks"
	"github.com/sirupsen/logrus"
)

var (
	_ proxy.Listener = (*Listener)(nil)
)

type Listener struct {
	*internal.TcpListener
}

func NewSocksListener() *Listener {
	return &Listener{
		TcpListener: internal.NewTcpListener("socks-listener", net.DefaultTcpOptions()),
	}
}

func (t *Listener) ProxyType() proxy.ProxyType {
	return proxy.ProxyType_SOCKS5
}

func (t *Listener) Serve(serveCtx context.Context, handler proxy.ListenerHandler) error {
	return t.TcpListener.Serve(serveCtx, func(connCtx context.Context, conn net.Connection) {
		if dest, err := socks.Handshake(conn); err != nil {
			logrus.Errorf("socks-listener: handshake: %s", err)
		} else if dest, err := parseSocksAddr(dest); err != nil {
			logrus.Errorf("socks-listener: parse destination: %s", err)
		} else {
			handler(connCtx, net.Connection{
				Network:     t.Network(),
				Address:     conn.Address,
				TCPConn:     conn.TCPConn,
				Destination: dest,
				ReadWriter:  conn.ReadWriter,
			})
		}
	})
}

func parseSocksAddr(target socks.Addr) (net.Destination, error) {
	switch target[0] {
	case socks.AtypDomainName:
		hAddr := string(target[2 : 2+target[1]])
		hPort := (int(target[2+target[1]]) << 8) | int(target[2+target[1]+1])
		if port, err := net.PortFromInt(uint32(hPort)); err != nil {
			return net.DestinationNotset, fmt.Errorf("invalid socks domain port: %d", hPort)
		} else {
			return net.Destination{
				Network: net.Network_TCP,
				Address: net.DomainAddress(hAddr),
				Port:    port,
			}, nil
		}

	case socks.AtypIPv4:
		//v4ip := net.IP(target[1 : 1+net.IPv4len])
		v4ip := net.IPAddress(target[1 : 1+net.IPv4len])
		v4port := (int(target[1+net.IPv4len]) << 8) | int(target[1+net.IPv4len+1])
		port, err := net.PortFromInt(uint32(v4port))
		if err != nil {
			return net.DestinationNotset, fmt.Errorf("invalid socks ipv4 port: %d", v4port)
		}
		return net.Destination{
			Network: net.Network_TCP,
			Address: v4ip,
			Port:    port,
		}, nil
	case socks.AtypIPv6:
		//v6ip := net.IP(target[1 : 1+net.IPv6len])
		v6ip := net.IPAddress(target[1 : 1+net.IPv6len])
		v6port := (int(target[1+net.IPv6len]) << 8) | int(target[1+net.IPv6len+1])
		port, err := net.PortFromInt(uint32(v6port))
		if err != nil {
			return net.DestinationNotset, fmt.Errorf("invalid socks ipv6 port: %d", v6port)
		}
		return net.Destination{
			Network: net.Network_TCP,
			Address: v6ip,
			Port:    port,
		}, nil
	}
	return net.DestinationNotset, fmt.Errorf("unsupported socks destination: %s", target)
}
