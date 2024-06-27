package fluxway

import (
	"context"
	"fluxway/helper"
	"fluxway/net"
	"fluxway/proxy"
	"fmt"
	"github.com/bytepowered/assert-go"
	"strings"
)

const (
	ServerModeMixin   string = "mixin"
	ServerModeProxy   string = "proxy"
	ServerModeForward string = "forward"
)

type ServerOptions struct {
	// Generic
	Mode     string `yaml:"mode"`
	AllowLan bool   `yaml:"allow_lan"`
	Bind     string `yaml:"bind"`
	// Http proxy server only
	HttpPort  int `yaml:"http_port"`
	HttpsPort int `yaml:"https_port"`
	// Socks proxy server only
	SocksPort int `yaml:"socks_port"`
}

type DispatchServer struct {
	opts     ServerOptions
	listener proxy.Listener
	router   proxy.Router
	resolver proxy.Resolver
	selector proxy.ConnectorSelector
}

func NewGenericServer(opts ServerOptions) *DispatchServer {
	assert.MustNotEmpty(opts.Mode, "server mode is required")
	return &DispatchServer{
		opts: opts,
	}
}

func (s *DispatchServer) Options() ServerOptions {
	return s.opts
}

func (s *DispatchServer) SetListener(listener proxy.Listener) {
	s.listener = listener
}

func (s *DispatchServer) SetRouter(router proxy.Router) {
	s.router = router
}

func (s *DispatchServer) SetResolver(resolver proxy.Resolver) {
	s.resolver = resolver
}

func (s *DispatchServer) SetConnector(c proxy.Connector) {
	s.SetConnectorSelector(func(conn *net.Connection) (proxy.Connector, bool) {
		return c, true
	})
}

func (s *DispatchServer) SetConnectorSelector(f proxy.ConnectorSelector) {
	s.selector = f
}

func (s *DispatchServer) Serve(servContext context.Context) error {
	assert.MustNotNil(s.listener, "server listener is nil")
	assert.MustNotNil(s.router, "server router is nil")
	assert.MustNotNil(s.selector, "server connector-selector is nil")
	return s.listener.Serve(servContext, func(connCtx context.Context, conn net.Connection) error {
		assert.MustTrue(connCtx != servContext, "server context must be new")
		_ = proxy.RequiredID(connCtx)
		connCtx = context.WithValue(connCtx, proxy.CtxKeyProxyType, s.listener.ProxyType())
		// Route
		routed, err := s.router.Route(connCtx, &conn)
		if err != nil {
			return fmt.Errorf("server route: %w", err)
		}
		destNetwork := routed.Destination.Network
		destAddr := routed.Destination.Address
		// ---- check route values
		assert.MustTrue(routed.Destination.IsValid(), "routed.Destination is invalid")
		if destNetwork == net.Network_TCP {
			assert.MustNotNil(routed.TCPConn, "routed.TCPConn is required")
			assert.MustNotNil(routed.ReadWriter, "routed.readWriter is nil")
		} else {
			assert.MustNil(routed.TCPConn, "routed.TCPConn must be nil")
		}
		// ---- resolve dest addr
		if destNetwork == net.Network_TCP || destNetwork == net.Network_UDP {
			if destAddr.Family().IsDomain() {
				if ip, err := s.resolver.Resolve(connCtx, destAddr.Domain()); err != nil {
					return fmt.Errorf("server resolve: %w", err)
				} else {
					routed.Destination.Address = net.IPAddress(ip)
				}
			}
		}
		// Connect
		connector, ok := s.selector(&routed)
		assert.MustTrue(ok, "connector not found, network: %s", destNetwork)
		if err := connector.DialServe(connCtx, &routed); err != nil && helper.IsConnectionClosed(err) {
			return nil
		} else {
			return err
		}
	})
}

func assertServerModeValid(mode string) {
	valid := false
	switch strings.ToLower(mode) {
	case ServerModeForward, ServerModeMixin, ServerModeProxy:
		valid = true
	default:
		valid = false
	}
	assert.MustTrue(valid, "invalid server mode: %s", mode)
}
