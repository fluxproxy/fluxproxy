package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/feature"
	"github.com/rocket-proxy/rocket-proxy/feature/authenticator"
	"github.com/rocket-proxy/rocket-proxy/feature/listener"
	"github.com/rocket-proxy/rocket-proxy/feature/ruleset"
	"github.com/rocket-proxy/rocket-proxy/helper"
	"github.com/rocket-proxy/rocket-proxy/net"
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
	listeners  []rocket.Listener
	dispatcher rocket.Dispatcher
	await      sync.WaitGroup
}

func NewApp() *App {
	return &App{
		await: sync.WaitGroup{},
	}
}

func (a *App) Init(runCtx context.Context, cmdMode string) error {
	// Server mode
	var serverConfig ServerConfig
	if err := unmarshalWith(runCtx, configPathServer, &serverConfig); err != nil {
		return err
	}
	if err := a.checkServerMode(serverConfig.Mode); err != nil {
		return fmt.Errorf("inst: %w", err)
	}
	forceChanged := false
	if cmdMode != RunServerModeAuto {
		forceChanged = serverConfig.Mode != cmdMode
		serverConfig.Mode = cmdMode
	}
	if forceChanged {
		logrus.Infof("inst: server mode: %s (force changed by command)", serverConfig.Mode)
	} else {
		logrus.Infof("inst: server mode: %s", serverConfig.Mode)
	}
	// Dispatcher
	a.dispatcher = feature.NewDispatcher(feature.DispatcherOptions{
		Verbose: serverConfig.Verbose,
	})
	if err := a.dispatcher.Init(runCtx); err != nil {
		return fmt.Errorf("inst: dispacher: %w", err)
	}
	// Resolver
	a.initResolver(runCtx)
	// Authenticator
	a.initAuthenticator(runCtx)
	// Ruleset
	a.initRuleset(runCtx)
	// Http listener
	if helper.ContainsAnyString(serverConfig.Mode, RunServerModeAuto, RunServerModeHttp) {
		if err := a.initHttpListener(runCtx, serverConfig); err != nil {
			return err
		}
	}
	// Socks listener
	if helper.ContainsAnyString(serverConfig.Mode, RunServerModeAuto, RunServerModeSocks) {
		if err := a.initSocksListener(runCtx, serverConfig); err != nil {
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

	servErrors := make(chan error, len(a.listeners)+1)
	// Dispatcher
	go func() {
		servErrors <- a.dispatcher.Serve(servCtx)
	}()
	// Listeners
	for _, srv := range a.listeners {
		a.await.Add(1)
		go func(lis rocket.Listener) {
			if err := lis.Listen(servCtx, a.dispatcher);
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

func (a *App) initHttpListener(runCtx context.Context, serverConfig ServerConfig) error {
	var httpConfig HttpConfig
	if err := unmarshalWith(runCtx, configPathServerHttp, &httpConfig); err != nil {
		return fmt.Errorf("inst: unmarshal http config. %w", err)
	}
	if httpConfig.Disabled {
		logrus.Warnf("inst: http server is disabled")
		return nil
	}
	if httpConfig.Port <= 0 {
		httpConfig.Port = 1080
	}
	if httpConfig.Bind == "" {
		httpConfig.Bind = "0.0.0.0"
	}
	var auth AuthenticatorConfig
	_ = unmarshalWith(runCtx, configPathAuthenticator, &auth)
	httpListener := listener.NewHttpListener(rocket.ListenerOptions{
		Address: httpConfig.Bind,
		Port:    httpConfig.Port,
		Verbose: serverConfig.Verbose,
		Auth:    auth.Enabled,
	}, listener.HttpOptions{})
	a.listeners = append(a.listeners, httpListener)
	return httpListener.Init(runCtx)
}

func (a *App) initSocksListener(runCtx context.Context, serverConfig ServerConfig) error {
	var socksConfig SocksConfig
	if err := unmarshalWith(runCtx, configPathServerSocks, &socksConfig); err != nil {
		return fmt.Errorf("inst: unmarshal socks config. %w", err)
	}
	if socksConfig.Disabled {
		logrus.Warnf("inst: socks server is disabled")
		return nil
	}
	if socksConfig.Port <= 0 {
		socksConfig.Port = 1081
	}
	if socksConfig.Bind == "" {
		socksConfig.Bind = "0.0.0.0"
	}
	var auth AuthenticatorConfig
	_ = unmarshalWith(runCtx, configPathAuthenticator, &auth)
	socksListener := listener.NewSocksListener(rocket.ListenerOptions{
		Address: socksConfig.Bind,
		Port:    socksConfig.Port,
		Verbose: serverConfig.Verbose,
		Auth:    auth.Enabled,
	}, listener.SocksOptions{})
	a.listeners = append(a.listeners, socksListener)
	return socksListener.Init(runCtx)
}

func (a *App) initResolver(runCtx context.Context) {
	var config ResolverConfig
	_ = unmarshalWith(runCtx, configPathResolver, &config)
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
	for name, userIP := range config.Hosts {
		rAddr, err := net.ParseAddress(net.NetworkTCP, userIP+":80")
		if err != nil {
			logrus.Warnf("resolver.hosts.%s=%s is not ip address", name, userIP)
		} else if rAddr.IsIP() {
			inst.Set(name, rAddr.IP)
		} else {
			logrus.Warnf("resolver.hosts.%s=%s is not ip address", name, userIP)
		}
	}
}

func (a *App) initAuthenticator(runCtx context.Context) {
	var config AuthenticatorConfig
	_ = unmarshalWith(runCtx, configPathAuthenticator, &config)
	if !config.Enabled {
		return
	}
	dispatcher := a.dispatcher.(*feature.Dispatcher)
	// Basic
	basic := authenticator.NewUsersAuthenticator(config.Basic)
	dispatcher.RegisterAuthenticator(rocket.AuthenticateBasic, basic)
}

func (a *App) initRuleset(runCtx context.Context) {
	var config []RulesetConfig
	_ = unmarshalWith(runCtx, configPathRuleset, &config)
	// 最高优先级：禁止回环访问
	rulesets := []rocket.Ruleset{
		ruleset.NewLoopback(loadLocalAddrs(runCtx)),
	}
	// builder
	buildIPNet := func(rule RulesetConfig) rocket.Ruleset {
		nets := make([]stdnet.IPNet, 0, len(rule.Address))
		for _, sAddr := range rule.Address {
			if _, ipNet, err := stdnet.ParseCIDR(sAddr); err == nil {
				nets = append(nets, *ipNet)
			} else {
				logrus.Warnf("invalid ruleset(ipnet) address: %s", sAddr)
			}
		}
		return ruleset.NewIPNet(strings.EqualFold(rule.Access, "allow"), strings.EqualFold(rule.Origin, "source"), nets)
	}
	// 第二优先级：其它规则
	for _, itemConfig := range config {
		switch strings.ToLower(itemConfig.Type) {
		case "ipnet":
			rulesets = append(rulesets, buildIPNet(itemConfig))
		}
	}
	feature.InitMultiRuleset(rulesets)
}

func (a *App) checkServerMode(mode string) error {
	switch strings.ToLower(mode) {
	case RunServerModeAuto, RunServerModeHttp, RunServerModeSocks:
		return nil
	default:
		return fmt.Errorf("invalid server mode: %s", mode)
	}
}
