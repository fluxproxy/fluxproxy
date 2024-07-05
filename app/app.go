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
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"sync"
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

func (i *App) Init(runCtx context.Context, cmdMode string) error {
	// Server mode
	var serverConfig ServerConfig
	if err := unmarshalWith(runCtx, configPathServer, &serverConfig); err != nil {
		return err
	}
	if err := i.checkServerMode(serverConfig.Mode); err != nil {
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
	i.dispatcher = feature.NewDispatcher()
	if err := i.dispatcher.Init(runCtx); err != nil {
		return fmt.Errorf("inst: dispacher: %w", err)
	}
	// Http listener
	if helper.ContainsAnyString(serverConfig.Mode, RunServerModeAuto, RunServerModeHttp) {
		if err := i.initHttpListener(runCtx); err != nil {
			return err
		}
	}
	// Socks listener

	// 初始化服务
	if len(i.listeners) == 0 {
		return fmt.Errorf("inst: no available listeners")
	}
	return nil
}

func (i *App) Serve(runCtx context.Context) error {
	servCtx, servCancel := context.WithCancel(runCtx)
	defer servCancel()

	servErrors := make(chan error, len(i.listeners)+1)
	// Dispatcher
	go func() {
		servErrors <- i.dispatcher.Serve(servCtx)
	}()
	// Listeners
	for _, srv := range i.listeners {
		i.await.Add(1)
		go func(lis rocket.Listener) {
			if err := lis.Listen(servCtx, i.dispatcher);
				err == nil ||
					errors.Is(err, context.Canceled) ||
					errors.Is(err, http.ErrServerClosed) {
				servErrors <- nil
			} else {
				servErrors <- err
			}
			i.await.Done()
		}(srv)
	}
	select {
	case err := <-servErrors:
		servCancel()
		return i.term(err)
	case <-runCtx.Done():
		servCancel()
		return i.term(nil)
	}
}

func (i *App) term(err error) error {
	i.await.Wait()
	return err
}

func (i *App) initHttpListener(runCtx context.Context) error {
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
	})
	i.listeners = append(i.listeners, inst)
	return inst.Init(runCtx)
}

func (i *App) checkServerMode(mode string) error {
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
