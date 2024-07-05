package feature

import (
	"context"
	"errors"
	"github.com/bytepowered/assert"
	"github.com/bytepowered/goes"
	"github.com/fluxproxy/fluxproxy/feature/authenticator"
	"github.com/fluxproxy/fluxproxy/feature/dialer"
	"github.com/fluxproxy/fluxproxy/helper"
	"github.com/fluxproxy/fluxproxy/internal"
	"github.com/fluxproxy/fluxproxy/net"
	"github.com/sirupsen/logrus"
	"math"
	"strings"
	"time"
)

var (
	_ proxy.Dispatcher = (*Dispatcher)(nil)
)

type DispatcherOptions struct {
	Verbose bool
}

type Dispatcher struct {
	opts          DispatcherOptions
	tunnels       chan proxy.Tunnel
	dialer        map[string]proxy.Dialer
	authenticator map[string]proxy.Authenticator
}

func NewDispatcher(opts DispatcherOptions) *Dispatcher {
	return &Dispatcher{
		opts:    opts,
		tunnels: make(chan proxy.Tunnel, math.MaxInt32),
	}
}

func (d *Dispatcher) Init(ctx context.Context) error {
	d.dialer = map[string]proxy.Dialer{
		dialer.DIRECT: dialer.NewTcpDirectDialer(),
		dialer.REJECT: dialer.NewRejectDialer(),
	}
	d.authenticator = map[string]proxy.Authenticator{
		proxy.AuthenticateAllow:  authenticator.NewAllowAuthenticator(),
		proxy.AuthenticateBearer: authenticator.NewDenyAuthenticator(),
		proxy.AuthenticateSource: authenticator.NewDenyAuthenticator(),
		proxy.AuthenticateToken:  authenticator.NewDenyAuthenticator(),
	}
	return nil
}

func (d *Dispatcher) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("dispatcher: serve:done")
			return ctx.Err()

		case v := <-d.tunnels:
			goes.Go(func() {
				d.handle(v)
			})
		}
	}
}

func (d *Dispatcher) Submit(s proxy.Tunnel) {
	d.tunnels <- s
}

func (d *Dispatcher) Authenticate(ctx context.Context, authentication proxy.Authentication) error {
	assert.MustTrue(authentication.Authenticate != proxy.AuthenticateAllow, "authenticate is invalid")
	auErr := d.lookupAuthenticator(authentication).Authenticate(ctx, authentication)
	if auErr != nil {
		proxy.Logger(ctx).Errorf("dispatcher: authenticate: %s", auErr)
	}
	return auErr
}

func (d *Dispatcher) RegisterAuthenticator(name string, authenticator proxy.Authenticator) {
	assert.MustNotEmpty(name, "authenticator name")
	name = strings.ToUpper(name)
	assert.MustFalse(strings.EqualFold(name, proxy.AuthenticateAllow), "authenticator name is invalid")
	assert.MustNotNil(authenticator, "authenticator is nil")
	_, exists := d.authenticator[name]
	assert.MustFalse(exists, "authenticator is already exists: %s", name)
	d.authenticator[name] = authenticator
	logrus.Infof("dispatcher: register:authenticator: %s", name)
}

func (d *Dispatcher) handle(local proxy.Tunnel) {
	defer helper.Close(local)
	destAddr := local.Destination()

	defer func(start time.Time) {
		proxy.Logger(local.Context()).
			WithField("duration", time.Since(start).String()).
			Infof("dispatcher: TERM")
	}(local.Context().Value(internal.CtxKeyStartTime).(time.Time))

	// Ruleset
	ruErr := UseRuleset().Allow(local.Context(), proxy.Permit{
		Source:      local.Source(),
		Destination: destAddr,
	})
	if hook, ok := internal.LookupHook(local.Context(), internal.CtxHookAfterRuleset); ok {
		if hErr := hook(local.Context(), ruErr); hErr != nil {
			proxy.Logger(local.Context()).Errorf("dispatcher: hook:ruleset: %s", hErr)
			return
		}
	}
	if ruErr != nil {
		if !errors.Is(ruErr, proxy.ErrNoRulesetMatched) {
			proxy.Logger(local.Context()).Errorf("dispatcher: ruleset: %s", ruErr)
			return
		}
	}

	// Resolve
	destIPAddr, rvErr := UseResolver().Resolve(local.Context(), destAddr)
	if rvErr != nil {
		proxy.Logger(local.Context()).Errorf("dispatcher: resolve: %s", rvErr)
		return
	}

	// Dial
	if d.opts.Verbose {
		proxy.Logger(local.Context()).
			WithField("ipaddr", destIPAddr.String()).
			Infof("dispatcher: DIAL")
	}
	remote, dlErr := d.lookupDialer(destAddr).Dial(local.Context(), net.Address{
		Network: destAddr.Network,
		Family:  net.ToAddressFamily(destIPAddr),
		IP:      destIPAddr,
		Port:    destAddr.Port,
	})
	defer helper.Close(remote)
	if hook, ok := internal.LookupHook(local.Context(), internal.CtxHookAfterDialed); ok {
		if hErr := hook(local.Context(), dlErr); hErr != nil {
			proxy.Logger(local.Context()).Errorf("dispatcher: hook:dial: %s", hErr)
			return
		}
	}
	if dlErr != nil {
		proxy.Logger(local.Context()).Errorf("dispatcher: dial: %s", dlErr)
		return
	}

	// Connect
	tErr := local.Connect(remote)
	d.onTailError(local.Context(), tErr)

}

func (d *Dispatcher) lookupDialer(addr net.Address) proxy.Dialer {
	return d.dialer[dialer.DIRECT]
}

func (d *Dispatcher) lookupAuthenticator(auth proxy.Authentication) proxy.Authenticator {
	v, ok := d.authenticator[auth.Authenticate]
	if !ok {
		return d.authenticator[proxy.AuthenticateAllow]
	}
	return v
}

func (d *Dispatcher) onTailError(connCtx context.Context, tErr error) {
	if tErr == nil {
		return
	}
	if !helper.IsCopierError(tErr) && !errors.Is(tErr, context.Canceled) {
		internal.LogTailError(connCtx, "http", tErr)
	}
}
