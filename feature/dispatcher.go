package feature

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/fluxproxy/fluxproxy"
	"github.com/fluxproxy/fluxproxy/feature/authenticator"
	"github.com/fluxproxy/fluxproxy/feature/dialer"
	"github.com/fluxproxy/fluxproxy/helper"
	"github.com/fluxproxy/fluxproxy/internal"
	"github.com/fluxproxy/fluxproxy/net"
	"github.com/sirupsen/logrus"
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
	dialer        map[string]proxy.Dialer
	authenticator map[proxy.Authenticate]proxy.Authenticator
}

func NewDispatcher(opts DispatcherOptions) *Dispatcher {
	return &Dispatcher{
		opts: opts,
	}
}

func (d *Dispatcher) Init(ctx context.Context) error {
	d.dialer = map[string]proxy.Dialer{
		dialer.DIRECT: dialer.NewTcpDirectDialer(),
		dialer.REJECT: dialer.NewRejectDialer(),
	}
	d.authenticator = map[proxy.Authenticate]proxy.Authenticator{
		proxy.AuthenticateAllow:  authenticator.NewAllowAuthenticator(),
		proxy.AuthenticateBearer: authenticator.NewDenyAuthenticator(),
		proxy.AuthenticateSource: authenticator.NewDenyAuthenticator(),
		proxy.AuthenticateToken:  authenticator.NewDenyAuthenticator(),
	}
	return nil
}

func (d *Dispatcher) Dispatch(local proxy.Connector) {
	defer helper.Close(local)
	destAddr := local.Destination()

	defer func(start time.Time) {
		proxy.Logger(local.Context()).
			WithField("duration", time.Since(start).String()).
			Infof("disp: term")
	}(local.Context().Value(internal.CtxKeyStartTime).(time.Time))

	// Ruleset
	ruErr := UseRuleset().Allow(local.Context(), proxy.Permit{
		Source:      local.Source(),
		Destination: destAddr,
	})
	ruErr = d.callHook(local, internal.CtxHookAfterRuleset, ruErr, "ruleset")
	if ruErr != nil {
		if !errors.Is(ruErr, proxy.ErrNoRulesetMatched) {
			proxy.Logger(local.Context()).Errorf("disp: ruleset: %s", ruErr)
			return
		}
	}

	// Resolve
	destIPAddr, rvErr := UseResolver().Resolve(local.Context(), destAddr)
	rvErr = d.callHook(local, internal.CtxHookAfterResolve, rvErr, "resolve")
	if rvErr != nil {
		proxy.Logger(local.Context()).Errorf("disp: resolve: %s", rvErr)
		return
	}

	// Dial
	if d.opts.Verbose {
		proxy.Logger(local.Context()).
			WithField("ipaddr", destIPAddr.String()).
			Infof("disp: dial")
	}
	remote, dlErr := d.lookupDialer(destAddr).Dial(local.Context(), net.Address{
		Network: destAddr.Network,
		Family:  net.ToAddressFamily(destIPAddr),
		IP:      destIPAddr,
		Port:    destAddr.Port,
	})
	defer helper.Close(remote)
	dlErr = d.callHook(local, internal.CtxHookAfterDial, dlErr, "dial")
	if dlErr != nil {
		proxy.Logger(local.Context()).Errorf("disp: dial: %s", dlErr)
		return
	}

	// Connect
	cnErr := local.Connect(remote)
	cnErr = d.callHook(local, internal.CtxHookAfterConnect, cnErr, "connect")
	d.onTailError(local.Context(), cnErr)
}

func (d *Dispatcher) Authenticate(ctx context.Context, authentication proxy.Authentication) error {
	assert.MustTrue(authentication.Authenticate != proxy.AuthenticateAllow, "authenticate is invalid")
	auErr := d.lookupAuthenticator(authentication).Authenticate(ctx, authentication)
	if auErr != nil {
		proxy.Logger(ctx).Errorf("disp: authenticate: %s", auErr)
	}
	return auErr
}

func (d *Dispatcher) RegisterAuthenticator(kind proxy.Authenticate, authenticator proxy.Authenticator) {
	assert.MustFalse(kind == proxy.AuthenticateAllow, "authenticator kind is invalid")
	assert.MustNotNil(authenticator, "authenticator is nil")
	_, exists := d.authenticator[kind]
	assert.MustFalse(exists, "authenticator is already exists: %s", kind)
	d.authenticator[kind] = authenticator
	logrus.Infof("disp: register:authenticator: %s", kind)
}

func (d *Dispatcher) lookupDialer(addr net.Address) proxy.Dialer {
	return d.dialer[dialer.DIRECT]
}

func (d *Dispatcher) callHook(local proxy.Connector, hookKey any, preErr error, phase string) error {
	if hook, ok := local.HookFunc(hookKey); ok {
		if hErr := hook(local.Context(), preErr); hErr != nil {
			proxy.Logger(local.Context()).Errorf("disp: hook:%s: %s", phase, hErr)
			if preErr != nil {
				return fmt.Errorf("%w . call %s hook. %w", preErr, phase, hErr)
			} else {
				return fmt.Errorf("call %s hook. %w", phase, hErr)
			}
		}
	}
	return preErr
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
		msg := tErr.Error()
		if strings.Contains(msg, "i/o timeout") {
			return
		}
		if strings.Contains(msg, "connection reset by peer") {
			return
		}
		proxy.Logger(connCtx).Errorf("disp: conn error: %s", tErr)
	}
}
