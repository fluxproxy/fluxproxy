package socks

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/internal"
	"github.com/rocketmanapp/rocket-proxy/modules/socks/v5"
	"github.com/rocketmanapp/rocket-proxy/net"
	"io"
	stdnet "net"
	"strings"
)

var (
	_ rocket.Listener = (*Listener)(nil)
)

type Listener struct {
	*internal.TcpListener
}

func NewSocksListener() *Listener {
	return &Listener{
		TcpListener: internal.NewTcpListener("socks", net.DefaultTcpOptions()),
	}
}

func (t *Listener) Listen(serveCtx context.Context, dispatchHandler rocket.ListenerHandler) error {
	return t.TcpListener.Listen(serveCtx, func(connCtx context.Context, conn net.Connection) error {
		return t.handle(connCtx, conn.TCPConn(), dispatchHandler)
	})
}

func (t *Listener) handle(connCtx context.Context, conn net.Conn, dispatchHandler rocket.ListenerHandler) error {
	bufConn := bufio.NewReader(conn)
	if method, err := v5.ParseMethodRequest(bufConn); err != nil {
		return err
	} else if method.Ver != v5.VersionSocks5 {
		return v5.ErrNotSupportVersion
	}
	request, err := v5.ParseRequest(bufConn)
	if err != nil {
		if errors.Is(err, v5.ErrUnrecognizedAddrType) {
			if err := send(conn, v5.RepAddrTypeNotSupported, nil); err != nil {
				return fmt.Errorf("failed to send reply %w", err)
			}
		}
		return fmt.Errorf("failed to read destination address, %w", err)
	}
	switch request.Command {
	case v5.CommandConnect:
		return t.handleConnect(connCtx, conn, request, dispatchHandler)
	case v5.CommandAssociate:
		return t.handleAssociate(connCtx, conn, request, dispatchHandler)
	case v5.CommandBind:
		return t.handleBind(connCtx, conn, request, dispatchHandler)
	default:
		return t.handleNotSupported(connCtx, conn, request)
	}
}

func (t *Listener) handleConnect(connCtx context.Context, conn net.Conn, r v5.Request, dispatchHandler rocket.ListenerHandler) error {
	// Send success
	if sErr := send(conn, v5.RepSuccess, conn.LocalAddr()); sErr != nil {
		return fmt.Errorf("socks send reply: %w", sErr)
	}
	// Next
	var destAddr net.Address
	if r.DstAddr.FQDN != "" {
		destAddr = net.DomainAddress(r.DstAddr.FQDN)
	} else {
		destAddr = net.IPAddress(r.DstAddr.IP)
	}
	hErr := dispatchHandler(connCtx, net.Connection{
		Network:     t.Network(),
		Address:     net.IPAddress((conn.RemoteAddr().(*stdnet.TCPAddr)).IP),
		ReadWriter:  conn.(*net.TCPConn),
		UserContext: context.Background(),
		Destination: net.Destination{
			Network: net.Network_TCP,
			Address: destAddr,
			Port:    net.Port(r.DstAddr.Port),
		},
	})
	// Complete
	if hErr != nil {
		msg := hErr.Error()
		resp := v5.RepHostUnreachable
		if strings.Contains(msg, "refused") {
			resp = v5.RepConnectionRefused
		} else if strings.Contains(msg, "network is unreachable") {
			resp = v5.RepNetworkUnreachable
		}
		if err := send(conn, resp, conn.LocalAddr()); err != nil {
			return fmt.Errorf("socks send reply, %v", err)
		}
		return hErr
	} else {
		return nil
	}
}

func (t *Listener) handleAssociate(connCtx context.Context, w io.Writer, r v5.Request, handler rocket.ListenerHandler) error {
	return t.handleNotSupported(connCtx, w, r)
}

func (t *Listener) handleBind(connCtx context.Context, w io.Writer, r v5.Request, _ rocket.ListenerHandler) error {
	return t.handleNotSupported(connCtx, w, r)
}

func (t *Listener) handleNotSupported(_ context.Context, w io.Writer, r v5.Request) error {
	if err := send(w, v5.RepCommandNotSupported, nil); err != nil {
		return fmt.Errorf("socks send reply: %w", err)
	}
	return fmt.Errorf("socks unsupported command: %v", r.Command)
}

func send(w io.Writer, rep uint8, bindAddr net.Addr) error {
	reply := v5.Reply{
		Version:  v5.VersionSocks5,
		Response: rep,
		BndAddr: v5.AddrSpec{
			AddrType: v5.ATYPIPv4,
			IP:       stdnet.IPv4zero,
			Port:     0,
		},
	}
	if reply.Response == v5.RepSuccess {
		if tcpAddr, ok := bindAddr.(*net.TCPAddr); ok && tcpAddr != nil {
			reply.BndAddr.IP = tcpAddr.IP
			reply.BndAddr.Port = tcpAddr.Port
		} else if udpAddr, ok := bindAddr.(*net.UDPAddr); ok && udpAddr != nil {
			reply.BndAddr.IP = udpAddr.IP
			reply.BndAddr.Port = udpAddr.Port
		} else {
			reply.Response = v5.RepAddrTypeNotSupported
		}

		if reply.BndAddr.IP.To4() != nil {
			reply.BndAddr.AddrType = v5.ATYPIPv4
		} else if reply.BndAddr.IP.To16() != nil {
			reply.BndAddr.AddrType = v5.ATYPIPv6
		}
	}
	_, err := w.Write(reply.Bytes())
	return err
}
