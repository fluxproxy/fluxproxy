package server

import (
	"context"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/net"
	stdnet "net"
	"strings"
	"time"
)

type Director struct {
	serverType rocket.ServerType
	serverOpts Options
	listener   rocket.Listener
	router     rocket.Router
	resolver   rocket.Resolver
	selector   rocket.ConnectorSelector
}

func NewDirector(opts Options) *Director {
	assert.MustNotEmpty(opts.Mode, "server mode is empty")
	return &Director{
		serverOpts: opts,
	}
}

func (d *Director) Options() Options {
	return d.serverOpts
}

func (d *Director) SetListener(listener rocket.Listener) {
	assert.MustNotNil(listener, "listener is nil")
	d.listener = listener
}

func (d *Director) SetRouter(router rocket.Router) {
	assert.MustNotNil(router, "router is nil")
	d.router = router
}

func (d *Director) SetResolver(resolver rocket.Resolver) {
	assert.MustNotNil(resolver, "resolver is nil")
	d.resolver = resolver
}

func (d *Director) SetConnector(c rocket.Connector) {
	assert.MustNotNil(c, "connector is nil")
	d.SetConnectorSelector(func(conn *net.Connection) (rocket.Connector, bool) {
		return c, true
	})
}

func (d *Director) SetConnectorSelector(f rocket.ConnectorSelector) {
	assert.MustNotNil(f, "connector-selector is nil")
	d.selector = f
}

func (d *Director) SetServerType(serverType rocket.ServerType) {
	d.serverType = serverType
}

func (d *Director) ServeListen(servContext context.Context) error {
	assert.MustNotNil(d.listener, "server listener is nil")
	assert.MustNotNil(d.router, "server router is nil")
	assert.MustNotNil(d.resolver, "server resolver is nil")
	assert.MustNotNil(d.selector, "server connector-selector is nil")
	return d.listener.Listen(servContext, func(connCtx context.Context, conn net.Connection) error {
		assert.MustTrue(connCtx != servContext, "conn context is the same ref as server context")
		assert.MustNotNil(conn.UserContext, "user context is nil")
		assert.MustNotEmpty(rocket.RequiredID(connCtx), "conn id is empty")
		if conn.Network == net.Network_TCP {
			_, isTcpConn := conn.ReadWriter.(*stdnet.TCPConn)
			assert.MustNotNil(isTcpConn, "conn read-writer is not type of *net.TCPConn")
		}

		defer func(start time.Time) {
			rocket.Logger(connCtx).Infof("%d: conn duration: %dms", d.serverType, time.Since(start).Milliseconds())
		}(time.Now())

		connCtx = context.WithValue(connCtx, rocket.CtxKeyServerType, d.serverType)
		routed, rErr := d.router.Route(connCtx, &conn)
		if rErr != nil {
			return fmt.Errorf("server router: %w", rErr)
		}
		destNetwork := routed.Destination.Network
		destAddr := routed.Destination.Address

		assert.MustTrue(routed.Destination.IsValid(), "routed destination is invalid")

		if ip, sErr := d.resolver.Resolve(connCtx, destAddr); sErr != nil {
			return fmt.Errorf("server resolve: %w", sErr)
		} else {
			routed.Destination.Address = net.IPAddress(ip)
		}

		connector, ok := d.selector(&routed)
		assert.MustTrue(ok, "connector not found, network: %d", destNetwork)
		if dErr := connector.DialServe(connCtx, &routed); dErr == nil {
			return nil
		} else {
			msg := dErr.Error()
			if strings.Contains(msg, "use of closed network connection") {
				return nil
			}
			if strings.Contains(msg, "i/o timeout") {
				return nil
			}
			if strings.Contains(msg, "connection reset by peer") {
				return nil
			}
			return dErr
		}
	})
}
