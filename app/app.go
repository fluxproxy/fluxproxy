package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/fluxproxy/fluxproxy"
	"github.com/fluxproxy/fluxproxy/feature"
	"github.com/fluxproxy/fluxproxy/feature/authenticator"
	"github.com/fluxproxy/fluxproxy/feature/listener"
	"github.com/fluxproxy/fluxproxy/feature/ruleset"
	"github.com/fluxproxy/fluxproxy/helper"
	"github.com/fluxproxy/fluxproxy/net"
	"github.com/sirupsen/logrus"
	stdnet "net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	RunServerModeAuto  string = "auto"
	RunServerModeHttp  string = "http"
	RunServerModeSocks string = "socks"
)

type App struct {
	listeners  []proxy.Listener
	dispatcher proxy.Dispatcher
	await      sync.WaitGroup
	// shared config
	authConfig   AuthenticatorConfig
	serverConfig ServerConfig
}

func NewApp() *App {
	return &App{
		await: sync.WaitGroup{},
	}
}

func (a *App) Init(runCtx context.Context, cmdMode string) error {
	// Shared config
	if err := unmarshalWith(runCtx, configPathServer, &a.serverConfig); err != nil {
		return err
	}
	if err := unmarshalWith(runCtx, configPathAuthenticator, &a.authConfig); err != nil {
		return err
	}
	// Server mode
	if err := a.checkServerMode(a.serverConfig.Mode); err != nil {
		return fmt.Errorf("inst: %w", err)
	}
	forceChanged := false
	if cmdMode != RunServerModeAuto {
		forceChanged = a.serverConfig.Mode != cmdMode
		a.serverConfig.Mode = cmdMode
	}
	if forceChanged {
		logrus.Infof("inst: server mode: %s (force changed by command)", a.serverConfig.Mode)
	} else {
		logrus.Infof("inst: server mode: %s", a.serverConfig.Mode)
	}
	// Dispatcher
	a.dispatcher = feature.NewDispatcher(feature.DispatcherOptions{
		Verbose: a.serverConfig.Verbose,
	})
	if err := a.dispatcher.Init(runCtx); err != nil {
		return fmt.Errorf("inst: dispacher: %w", err)
	}
	// Resolver
	if err := a.initResolver(runCtx); err != nil {
		return fmt.Errorf("inst: init resolver: %w", err)
	}
	// Authenticator
	if err := a.initAuthenticator(runCtx); err != nil {
		return fmt.Errorf("inst: init authenticator: %w", err)
	}
	// Ruleset
	if err := a.initRuleset(runCtx); err != nil {
		return fmt.Errorf("inst: init ruleset: %w", err)
	}
	// Http listener
	if helper.ContainsAny(a.serverConfig.Mode, RunServerModeAuto, RunServerModeHttp) {
		if err := a.initHttpListener(runCtx, a.dispatcher); err != nil {
			return err
		}
	}
	// Socks listener
	if helper.ContainsAny(a.serverConfig.Mode, RunServerModeAuto, RunServerModeSocks) {
		if err := a.initSocksListener(runCtx, a.dispatcher); err != nil {
			return err
		}
	}
	if len(a.listeners) == 0 {
		return fmt.Errorf("inst: no available listeners")
	}
	return nil
}

func (a *App) Serve(runCtx context.Context) error {
	servCtx, servCancel := context.WithCancel(runCtx)
	defer servCancel()
	servErrors := make(chan error, len(a.listeners))
	for _, srv := range a.listeners {
		a.await.Add(1)
		go func(lis proxy.Listener) {
			if err := lis.Listen(servCtx);
				err == nil ||
					errors.Is(err, context.Canceled) ||
					errors.Is(err, http.ErrServerClosed) {
				servErrors <- nil
			} else {
				servErrors <- err
			}
			a.await.Done()
		}(srv)
	}
	select {
	case err := <-servErrors:
		servCancel()
		return a.term(err)
	case <-runCtx.Done():
		servCancel()
		return a.term(nil)
	}
}

func (a *App) term(err error) error {
	a.await.Wait()
	return err
}

func (a *App) initHttpListener(runCtx context.Context, dispatcher proxy.Dispatcher) error {
	assert.MustNotNil(runCtx, "context is nil")
	assert.MustNotNil(dispatcher, "dispatcher is nil")
	var httpConfig HttpConfig
	if err := unmarshalWith(runCtx, configPathServerHttp, &httpConfig); err != nil {
		return fmt.Errorf("inst: unmarshal http config. %w", err)
	}
	if httpConfig.Disabled {
		logrus.Warnf("inst: http server is disabled")
		return nil
	}
	lstOpts := proxy.ListenerOptions{
		Address: convBindAddress(httpConfig.Bind),
		Port:    convBindPort(httpConfig.Port, 1080),
		Verbose: a.serverConfig.Verbose,
		Auth:    a.authConfig.Enabled,
	}
	httpOpts := listener.HttpOptions{}
	httpListener := listener.NewHttpListener(lstOpts, httpOpts, dispatcher)
	a.listeners = append(a.listeners, httpListener)
	return httpListener.Init(runCtx)
}

func (a *App) initSocksListener(runCtx context.Context, dispatcher proxy.Dispatcher) error {
	assert.MustNotNil(runCtx, "context is nil")
	assert.MustNotNil(dispatcher, "dispatcher is nil")
	var socksConfig SocksConfig
	if err := unmarshalWith(runCtx, configPathServerSocks, &socksConfig); err != nil {
		return fmt.Errorf("inst: unmarshal socks config. %w", err)
	}
	if socksConfig.Disabled {
		logrus.Warnf("inst: socks server is disabled")
		return nil
	}
	lstOpts := proxy.ListenerOptions{
		Address: convBindAddress(socksConfig.Bind),
		Port:    convBindPort(socksConfig.Port, 1081),
		Verbose: a.serverConfig.Verbose,
		Auth:    a.authConfig.Enabled,
	}
	socksOpts := listener.SocksOptions{}
	socksListener := listener.NewSocksListener(lstOpts, socksOpts, dispatcher)
	a.listeners = append(a.listeners, socksListener)
	return socksListener.Init(runCtx)
}

func (a *App) initResolver(runCtx context.Context) error {
	var config ResolverConfig
	if err := unmarshalWith(runCtx, configPathResolver, &config); err != nil {
		return fmt.Errorf("inst: unmarshal resolver config. %w", err)
	}
	if config.CacheSize <= 0 {
		config.CacheSize = 1024 * 10
	}
	if config.CacheTTL <= 0 {
		config.CacheTTL = 60
	}
	inst := feature.InitResolverWith(feature.Options{
		CacheSize: config.CacheSize,
		CacheTTL:  time.Duration(config.CacheTTL) * time.Second,
		Hosts:     config.Hosts,
	})
	// prepare
	for host, ipAddr := range config.Hosts {
		rIPAddr, err := net.ParseAddress(net.NetworkTCP, ipAddr+":80")
		if err != nil {
			return fmt.Errorf("resolver.hosts %s=%s is invalid", host, ipAddr)
		}
		if rIPAddr.IsIP() {
			inst.Set(host, rIPAddr.IP)
		} else {
			return fmt.Errorf("resolver.hosts %s=%s is not ip address", host, ipAddr)
		}
	}
	return nil
}

func (a *App) initAuthenticator(runCtx context.Context) error {
	if !a.authConfig.Enabled {
		logrus.Warnf("inst: authenticator is disabled")
		return nil
	}
	dispatcher := a.dispatcher.(*feature.Dispatcher)
	// Basic
	for u, p := range a.authConfig.Basic {
		if len(u) < 2 || len(p) < 2 {
			return fmt.Errorf("invalid user or password in authenticator.basic: %s=%s", u, p)
		}
	}
	basic := authenticator.NewUsersAuthenticator(a.authConfig.Basic)
	dispatcher.RegisterAuthenticator(proxy.AuthenticateBasic, basic)
	// Others todo
	return nil
}

func (a *App) initRuleset(runCtx context.Context) error {
	var config []RulesetConfig
	if err := unmarshalWith(runCtx, configPathRuleset, &config); err != nil {
		return fmt.Errorf("unmarshal ruleset. %w", err)
	}
	// builder
	ipnetBuilder := func(rule RulesetConfig) (proxy.Ruleset, error) {
		nets := make([]stdnet.IPNet, 0, len(rule.Address))
		for _, sAddr := range rule.Address {
			if _, ipNet, err := stdnet.ParseCIDR(sAddr); err == nil {
				nets = append(nets, *ipNet)
			} else {
				return nil, fmt.Errorf("invalid ruleset(ipnet) address: %s", sAddr)
			}
		}
		return ruleset.NewIPNet(strings.EqualFold(rule.Access, "allow"), strings.EqualFold(rule.Origin, "source"), nets), nil
	}
	// 最高优先级：禁止回环访问
	rulesets := []proxy.Ruleset{
		ruleset.NewLoopback(loadLocalAddrs(runCtx)),
	}
	// 第二优先级：其它规则
	for _, itemConfig := range config {
		switch strings.ToLower(itemConfig.Type) {
		case "ipnet":
			if ruleset, err := ipnetBuilder(itemConfig); err != nil {
				return err
			} else {
				rulesets = append(rulesets, ruleset)
			}
		}
	}
	feature.InitMultiRuleset(rulesets)
	return nil
}

func (a *App) checkServerMode(mode string) error {
	switch strings.ToLower(mode) {
	case RunServerModeAuto, RunServerModeHttp, RunServerModeSocks:
		return nil
	default:
		return fmt.Errorf("invalid server mode: %s", mode)
	}
}

func convBindAddress(bind string) string {
	if bind == "" {
		return "0.0.0.0"
	}
	return bind
}

func convBindPort(port int, def int) int {
	if port == 0 {
		return def
	}
	return port
}
