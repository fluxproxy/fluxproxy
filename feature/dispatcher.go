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
)

var (
	_ rocket.Dispatcher = (*Dispatcher)(nil)
)

type Dispatcher struct {
	tunnels       chan rocket.Tunnel
	dialer        map[string]rocket.Dialer
	authenticator map[string]rocket.Authenticator
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
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
	// Authenticate
	assert.MustTrue(local.Authentication().Authenticate != rocket.AuthenticateAllow, "authenticate is invalid")
	aErr := d.lookupAuthenticator(local.Authentication()).Authenticate(local.Context(), local.Authentication())
	// Hook: authed
	if hook, ok := internal.LookupHook(local.Context(), internal.CtxHookAfterAuthed); ok {
		if hErr := hook(local.Context(), aErr); hErr != nil {
			rocket.Logger(local.Context()).Errorf("dispatcher: hook:auth: %s", hErr)
			return
		}
	}
	destAddr := local.Destination()
	// Resolve
	destIPAddr, rErr := UseResolver().Resolve(local.Context(), destAddr)
	if rErr != nil {
		rocket.Logger(local.Context()).Errorf("dispatcher: resolve: %s", rErr)
		return
	}
	// Dial
	remote, dErr := d.lookupDialer(destAddr).Dial(local.Context(), net.Address{
		Network: destAddr.Network,
		Family:  net.ToAddressFamily(destIPAddr),
		IP:      destIPAddr,
		Port:    destAddr.Port,
	})
	if dErr != nil {
		rocket.Logger(local.Context()).Errorf("dispatcher: dial: %s", dErr)
		return
	}
	defer helper.Close(remote)
	// Hook: dialed
	if hook, ok := internal.LookupHook(local.Context(), internal.CtxHookAfterDialed); ok {
		if hErr := hook(local.Context(), nil); hErr != nil {
			rocket.Logger(local.Context()).Errorf("dispatcher: hook:dial: %s", hErr)
			return
		}
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
