package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/net"
	stdnet "net"
	"time"
)

type Director struct {
	serverType        rocket.ServerType
	serverConfig      ServerConfig
	listener          rocket.Listener
	router            rocket.Router
	resolver          rocket.Resolver
	ruleset           rocket.Ruleset
	connectorSelector rocket.ConnectorSelectFunc
	authenticator     rocket.Authenticator
}

func NewDirector(serverConfig ServerConfig) *Director {
	assert.MustNotEmpty(serverConfig.Mode, "server mode is empty")
	return &Director{
		serverConfig: serverConfig,
	}
}

func (d *Director) ServerConfig() ServerConfig {
	return d.serverConfig
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

func (d *Director) SetRuleset(ruleset rocket.Ruleset) {
	assert.MustNotNil(ruleset, "ruleset is nil")
	d.ruleset = ruleset
}

func (d *Director) SetConnector(c rocket.Connector) {
	assert.MustNotNil(c, "connector is nil")
	d.SetConnectorSelector(func(conn *net.Connection) (rocket.Connector, bool) {
		return c, true
	})
}

func (d *Director) SetConnectorSelector(f rocket.ConnectorSelectFunc) {
	assert.MustNotNil(f, "connector-selector is nil")
	d.connectorSelector = f
}

func (d *Director) SetAuthenticator(f rocket.Authenticator) {
	assert.MustNotNil(f, "authenticator is nil")
	d.authenticator = f
}

func (d *Director) SetServerType(serverType rocket.ServerType) {
	d.serverType = serverType
}

func (d *Director) ServeListen(servContext context.Context) error {
	assert.MustNotNil(d.listener, "server listener is nil")
	assert.MustNotNil(d.router, "server router is nil")
	assert.MustNotNil(d.resolver, "server resolver is nil")
	assert.MustNotNil(d.authenticator, "server authenticator is nil")
	assert.MustNotNil(d.ruleset, "server ruleset is nil")
	assert.MustNotNil(d.connectorSelector, "server connector-selector is nil")
	return d.listener.Listen(servContext, &rocket.ListenerHandlerAdapter{
		Authenticator: d.authenticator,
		Dispatcher: func(connCtx context.Context, conn net.Connection) error {
			assert.MustTrue(connCtx != servContext, "conn context is the same ref as server context")
			assert.MustNotNil(conn.UserContext, "user context is nil")
			assert.MustNotEmpty(rocket.RequiredID(connCtx), "conn id is empty")
			if conn.Network == net.NetworkTCP {
				_, isTcpConn := conn.ReadWriter.(*stdnet.TCPConn)
				assert.MustNotNil(isTcpConn, "conn read-writer is not type of *net.TCPConn")
			}

			defer func(start time.Time) {
				rocket.Logger(connCtx).Infof("%s: conn duration: %dms", d.serverType, time.Since(start).Milliseconds())
			}(time.Now())

			connCtx = context.WithValue(connCtx, rocket.CtxKeyServerType, d.serverType)
			connCtx, newConn, roErr := d.router.Route(connCtx, &conn)
			if roErr != nil {
				return fmt.Errorf("director: route. %w", roErr)
			} else {
				assert.MustTrue(newConn.Destination.IsValid(), "router destination is invalid")
				assert.MustNotNil(connCtx, "router dest context is nil")
			}

			if ip, reErr := d.resolver.Resolve(connCtx, newConn.Destination.Address); reErr != nil {
				return fmt.Errorf("director: resolve. %w", reErr)
			} else {
				newConn.Destination.Address = net.IPAddress(ip)
			}

			connCtx, rsErr := d.ruleset.Allow(connCtx, rocket.Permit{
				Source:      newConn.Address,
				Destination: newConn.Destination,
			})
			if rsErr != nil && !errors.Is(rsErr, rocket.ErrRulesetNotMatched) {
				return fmt.Errorf("director: ruleset. %w", rsErr)
			} else {
				assert.MustNotNil(connCtx, "ruleset dest context is nil")
			}

			connector, ok := d.connectorSelector(&newConn)
			assert.MustTrue(ok, "connector not found, network: %d", newConn.Destination.Network)
			return connector.DialServe(connCtx, &newConn)
		},
	})
}
