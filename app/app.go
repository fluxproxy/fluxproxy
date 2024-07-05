package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/knadh/koanf/v2"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/feature"
	"github.com/rocket-proxy/rocket-proxy/feature/listener"
	"github.com/rocket-proxy/rocket-proxy/helper"
	"github.com/rocket-proxy/rocket-proxy/net"
	"github.com/sirupsen/logrus"
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
	// Resolver
	a.initResolver(runCtx)

	// Dispatcher
	a.dispatcher = feature.NewDispatcher()
	if err := a.dispatcher.Init(runCtx); err != nil {
		return fmt.Errorf("inst: dispacher: %w", err)
	}
	// Http listener
	if helper.ContainsAnyString(serverConfig.Mode, RunServerModeAuto, RunServerModeHttp) {
		if err := a.initHttpListener(runCtx, serverConfig); err != nil {
			return err
		}
	}
	// Socks listener

	// 初始化服务
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
	if err := unmarshalWith(runCtx, configPathHttp, &httpConfig); err != nil {
		return fmt.Errorf("inst: unmarshal http config. %w", err)
	}
	if httpConfig.Disabled {
		logrus.Warnf("inst: http server is disabled")
		return nil
	}
	inst := listener.NewHttpListener(rocket.ListenerOptions{
		Address: httpConfig.Bind,
		Port:    httpConfig.Port,
	}, listener.HttpOptions{
		Verbose: serverConfig.Verbose,
	})
	a.listeners = append(a.listeners, inst)
	return inst.Init(runCtx)
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

func (a *App) checkServerMode(mode string) error {
	switch strings.ToLower(mode) {
	case RunServerModeAuto, RunServerModeHttp, RunServerModeSocks:
		return nil
	default:
		return fmt.Errorf("invalid server mode: %s", mode)
	}
}

func unmarshalWith(ctx context.Context, path string, out any) error {
	if err := rocket.Configer(ctx).UnmarshalWithConf(path, out, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return fmt.Errorf("config unmarshal %s. %w", path, err)
	}
	return nil
}
