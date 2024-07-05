package listener

import (
	"context"
	"errors"
	"fmt"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/feature/tunnel"
	"github.com/rocket-proxy/rocket-proxy/internal"
	"github.com/rocket-proxy/rocket-proxy/net"
	"github.com/rocket-proxy/rocket-proxy/statute/socks"
	"github.com/sirupsen/logrus"
	stdnet "net"
	"strconv"
	"strings"
)

var (
	_ rocket.Listener = (*SocksListener)(nil)
)

type SocksOptions struct {
}

type SocksListener struct {
	opts         SocksOptions
	listenerOpts rocket.ListenerOptions
}

func NewSocksListener(listenerOpts rocket.ListenerOptions, socksOpts SocksOptions) *SocksListener {
	return &SocksListener{
		listenerOpts: listenerOpts,
		opts:         socksOpts,
	}
}

func (l *SocksListener) Init(ctx context.Context) error {
	if l.listenerOpts.Port <= 0 {
		return fmt.Errorf("socks: invalid port: %d", l.listenerOpts.Port)
	}
	return nil
}

func (l *SocksListener) Listen(serveCtx context.Context, dispatcher rocket.Dispatcher) error {
	addr := stdnet.JoinHostPort(l.listenerOpts.Address, strconv.Itoa(l.listenerOpts.Port))
	if l.listenerOpts.Auth {
		logrus.Infof("socks: listen: %s", addr)
	} else {
		logrus.Infof("socks: listen(no auth): %s", addr)
	}
	return tcpListenWith(serveCtx, l.listenerOpts, func(tcpConn *stdnet.TCPConn) {
		connCtx := internal.SetupTcpContextLogger(serveCtx, tcpConn)
		if err := l.handshakeHeader(connCtx, tcpConn); err != nil {
			_ = l.send(tcpConn, socks.RepConnectionRefused)
			rocket.Logger(connCtx).Errorf("socks: header: %s", err)
			return
		}
		srcAddr := parseRemoteAddress(tcpConn.RemoteAddr().String())

		// Authenticate
		if l.listenerOpts.Auth {
			if err := l.handshakeUserAuth(connCtx, tcpConn, dispatcher); err != nil {
				rocket.Logger(connCtx).Errorf("socks: auth(user): %s", err)
				return
			}
		} else {
			if err := l.handshakeSkipAuth(connCtx, tcpConn, dispatcher); err != nil {
				rocket.Logger(connCtx).Errorf("socks: auth(skip): %s", err)
				return
			}
		}

		// Destination
		request, prErr := socks.ParseRequest(tcpConn)
		if prErr != nil {
			_ = l.send(tcpConn, socks.RepAddrTypeNotSupported)
			rocket.Logger(connCtx).Errorf("socks: auth(skip): %s", prErr)
			return
		}
		if request.Command != socks.CommandConnect {
			_ = l.send(tcpConn, socks.RepCommandNotSupported)
			return
		}
		var destAddr net.Address
		if request.DstAddr.FQDN != "" {
			destAddr = net.ParseDomainAddr(net.NetworkTCP, request.DstAddr.FQDN)
		} else {
			destAddr = net.ParseIPAddr(net.NetworkTCP, request.DstAddr.IP)
		}
		destAddr.Port = request.DstAddr.Port

		// Submit
		connCtx = internal.ContextWithHooks(connCtx, map[any]rocket.HookFunc{
			internal.CtxHookAfterRuleset: l.withRulesetHook(tcpConn),
			internal.CtxHookAfterDialed:  l.withDialedHook(tcpConn),
		})
		stream := tunnel.NewConnStream(connCtx, tcpConn, destAddr, srcAddr)
		dispatcher.Submit(stream)

		if l.listenerOpts.Verbose {
			rocket.Logger(connCtx).WithField("dest", request.DstAddr.String()).Infof("socks: CONN")
		}

	})
}

func (l *SocksListener) handshakeHeader(ctx context.Context, conn stdnet.Conn) error {
	if request, err := socks.ParseMethodRequest(conn); err != nil {
		return fmt.Errorf("parse method request. %w", err)
	} else if request.Ver != socks.VersionSocks5 {
		return socks.ErrNotSupportVersion
	}
	return nil
}

func (l *SocksListener) handshakeSkipAuth(ctx context.Context, conn stdnet.Conn, dispatcher rocket.Dispatcher) error {
	_, err := conn.Write([]byte{socks.VersionSocks5, socks.MethodNoAuth})
	return err
}

func (l *SocksListener) handshakeUserAuth(ctx context.Context, conn stdnet.Conn, dispatcher rocket.Dispatcher) error {
	if _, err := conn.Write([]byte{socks.VersionSocks5, socks.MethodUserPassAuth}); err != nil {
		return fmt.Errorf("send auth request. %w", err)
	}
	request, upErr := socks.ParseUserPassRequest(conn)
	if upErr != nil {
		return fmt.Errorf("parse auth request. %w", upErr)
	}
	auErr := dispatcher.Authenticate(ctx, rocket.Authentication{
		Source:         parseRemoteAddress(conn.RemoteAddr().String()),
		Authenticate:   rocket.AuthenticateBasic,
		Authentication: string(request.User) + ":" + string(request.Pass),
	})
	if auErr != nil {
		if _, err := conn.Write([]byte{socks.UserPassAuthVersion, socks.AuthFailure}); err != nil {
			return fmt.Errorf("send failed auth reply. %w", err)
		}
	} else {
		if _, err := conn.Write([]byte{socks.UserPassAuthVersion, socks.AuthSuccess}); err != nil {
			return fmt.Errorf("send success auth reply. %w", err)
		}
	}
	return auErr
}

func (l *SocksListener) withAuthorizedHook(conn stdnet.Conn) rocket.HookFunc {
	return func(ctx context.Context, state error, v ...any) error {
		var status byte
		if state == nil {
			status = socks.AuthSuccess
		} else {
			status = socks.AuthFailure
		}
		_, err := conn.Write([]byte{socks.UserPassAuthVersion, status})
		return err
	}
}

func (l *SocksListener) withRulesetHook(conn stdnet.Conn) rocket.HookFunc {
	return func(ctx context.Context, state error, v ...any) error {
		if state == nil || errors.Is(state, rocket.ErrNoRulesetMatched) {
			return nil
		}
		return l.send(conn, socks.RepRuleFailure)
	}
}

func (l *SocksListener) withDialedHook(conn stdnet.Conn) rocket.HookFunc {
	return func(_ context.Context, state error, _ ...any) error {
		if state == nil {
			return l.send(conn, socks.RepSuccess)
		}
		msg := state.Error()
		if strings.Contains(msg, "connection refused") {
			return l.send(conn, socks.RepConnectionRefused)
		} else if strings.Contains(msg, "network is unreachable") {
			return l.send(conn, socks.RepNetworkUnreachable)
		} else {
			return l.send(conn, socks.RepHostUnreachable)
		}
	}
}

func (*SocksListener) send(conn stdnet.Conn, rep uint8) error {
	bindAddr := conn.LocalAddr()
	reply := socks.Reply{
		Version:  socks.VersionSocks5,
		Response: rep,
		BndAddr: socks.AddrSpec{
			AddrType: socks.ATYPIPv4,
			IP:       stdnet.IPv4zero,
			Port:     0,
		},
	}
	if reply.Response == socks.RepSuccess {
		if tcpAddr, ok := bindAddr.(*stdnet.TCPAddr); ok && tcpAddr != nil {
			reply.BndAddr.IP = tcpAddr.IP
			reply.BndAddr.Port = tcpAddr.Port
		} else if udpAddr, ok := bindAddr.(*stdnet.UDPAddr); ok && udpAddr != nil {
			reply.BndAddr.IP = udpAddr.IP
			reply.BndAddr.Port = udpAddr.Port
		} else {
			reply.Response = socks.RepAddrTypeNotSupported
		}

		if reply.BndAddr.IP.To4() != nil {
			reply.BndAddr.AddrType = socks.ATYPIPv4
		} else if reply.BndAddr.IP.To16() != nil {
			reply.BndAddr.AddrType = socks.ATYPIPv6
		}
	}
	_, err := conn.Write(reply.Bytes())
	return err
}
