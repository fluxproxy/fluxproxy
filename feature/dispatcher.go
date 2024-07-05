package feature

import (
	"context"
	"errors"
	"github.com/bytepowered/assert"
	"github.com/bytepowered/goes"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/feature/authenticator"
	"github.com/rocket-proxy/rocket-proxy/feature/dialer"
	"github.com/rocket-proxy/rocket-proxy/helper"
	"github.com/rocket-proxy/rocket-proxy/internal"
	"github.com/rocket-proxy/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	"math"
	"strings"
	"time"
)

var (
	_ rocket.Dispatcher = (*Dispatcher)(nil)
)

type DispatcherOptions struct {
	Verbose bool
}

type Dispatcher struct {
	opts          DispatcherOptions
	tunnels       chan rocket.Tunnel
	dialer        map[string]rocket.Dialer
	authenticator map[string]rocket.Authenticator
}

func NewDispatcher(opts DispatcherOptions) *Dispatcher {
	return &Dispatcher{
		opts:    opts,
		tunnels: make(chan rocket.Tunnel, math.MaxInt32),
	}
}

func (d *Dispatcher) Init(ctx context.Context) error {
	d.dialer = map[string]rocket.Dialer{
		dialer.DIRECT: dialer.NewTcpDirectDialer(),
		dialer.REJECT: dialer.NewRejectDialer(),
	}
	d.authenticator = map[string]rocket.Authenticator{
		rocket.AuthenticateAllow:  authenticator.NewAllowAuthenticator(),
		rocket.AuthenticateBearer: authenticator.NewDenyAuthenticator(),
		rocket.AuthenticateSource: authenticator.NewDenyAuthenticator(),
		rocket.AuthenticateToken:  authenticator.NewDenyAuthenticator(),
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

func (d *Dispatcher) Submit(s rocket.Tunnel) {
	d.tunnels <- s
}

func (d *Dispatcher) Authenticate(ctx context.Context, authentication rocket.Authentication) error {
	assert.MustTrue(authentication.Authenticate != rocket.AuthenticateAllow, "authenticate is invalid")
	auErr := d.lookupAuthenticator(authentication).Authenticate(ctx, authentication)
	if auErr != nil {
		rocket.Logger(ctx).Errorf("dispatcher: authenticate: %s", auErr)
	}
	return auErr
}

func (d *Dispatcher) RegisterAuthenticator(name string, authenticator rocket.Authenticator) {
	assert.MustNotEmpty(name, "authenticator name")
	name = strings.ToUpper(name)
	assert.MustFalse(strings.EqualFold(name, rocket.AuthenticateAllow), "authenticator name is invalid")
	assert.MustNotNil(authenticator, "authenticator is nil")
	_, exists := d.authenticator[name]
	assert.MustFalse(exists, "authenticator is already exists: %s", name)
	d.authenticator[name] = authenticator
	logrus.Infof("dispatcher: register:authenticator: %s", name)
}

func (d *Dispatcher) handle(local rocket.Tunnel) {
	defer helper.Close(local)
	destAddr := local.Destination()

	defer func(start time.Time) {
		rocket.Logger(local.Context()).
			WithField("duration", time.Since(start).String()).
			Infof("dispatcher: TERM")
	}(local.Context().Value(internal.CtxKeyStartTime).(time.Time))

	// Ruleset
	ruErr := UseRuleset().Allow(local.Context(), rocket.Permit{
		Source:      local.Source(),
		Destination: destAddr,
	})
	if hook, ok := internal.LookupHook(local.Context(), internal.CtxHookAfterRuleset); ok {
		if hErr := hook(local.Context(), ruErr); hErr != nil {
			rocket.Logger(local.Context()).Errorf("dispatcher: hook:ruleset: %s", hErr)
			return
		}
	}
	if ruErr != nil {
		if !errors.Is(ruErr, rocket.ErrNoRulesetMatched) {
			rocket.Logger(local.Context()).Errorf("dispatcher: ruleset: %s", ruErr)
			return
		}
	}

	// Resolve
	destIPAddr, rvErr := UseResolver().Resolve(local.Context(), destAddr)
	if rvErr != nil {
		rocket.Logger(local.Context()).Errorf("dispatcher: resolve: %s", rvErr)
		return
	}

	// Dial
	if d.opts.Verbose {
		rocket.Logger(local.Context()).
			WithField("ip", destIPAddr.String()).
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
			rocket.Logger(local.Context()).Errorf("dispatcher: hook:dial: %s", hErr)
			return
		}
	}
	if dlErr != nil {
		rocket.Logger(local.Context()).Errorf("dispatcher: dial: %s", dlErr)
		return
	}

	// Connect
	tErr := local.Connect(remote)
	d.onTailError(local.Context(), tErr)

}

func (d *Dispatcher) lookupDialer(addr net.Address) rocket.Dialer {
	return d.dialer[dialer.DIRECT]
}

func (d *Dispatcher) lookupAuthenticator(auth rocket.Authentication) rocket.Authenticator {
	v, ok := d.authenticator[auth.Authenticate]
	if !ok {
		return d.authenticator[rocket.AuthenticateAllow]
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
