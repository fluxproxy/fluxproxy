package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/helper"
	"github.com/rocket-proxy/rocket-proxy/modules/listener"
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

type Instance struct {
	listeners  []rocket.Listener
	dispatcher rocket.Dispatcher
	await      sync.WaitGroup
}

func NewInstance() *Instance {
	return &Instance{
		await: sync.WaitGroup{},
	}
}

func (i *Instance) Init(runCtx context.Context, cmdMode string) error {
	// Server mode
	var serverConfig ServerConfig
	if err := rocket.ConfigerUnmarshal(runCtx, configPathServer, &serverConfig); err != nil {
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
	i.dispatcher = NewDispatcher()
	return nil
}

func (i *Instance) Serve(runCtx context.Context) error {
	servCtx, servCancel := context.WithCancel(runCtx)
	servErrors := make(chan error, len(i.listeners))
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

func (i *Instance) term(err error) error {
	i.await.Wait()
	return err
}

func (i *Instance) initHttpListener(runCtx context.Context) error {
	var httpConfig HttpConfig
	if err := rocket.ConfigerUnmarshal(runCtx, configPathHttp, &httpConfig); err != nil {
		return fmt.Errorf("inst: unmarshal http config. %w", err)
	}
	if httpConfig.Disabled {
		logrus.Warnf("inst: http server is disabled")
		return nil
	}
	if httpConfig.Port <= 0 {
		return fmt.Errorf("inst: invalid http port: %d", httpConfig.Port)
	}
	inst := listener.NewHttpListener(rocket.ListenerOptions{
		Address: httpConfig.Bind,
		Port:    httpConfig.Port,
	})
	i.listeners = append(i.listeners, inst)
	return inst.Init(runCtx)
}

func (i *Instance) checkServerMode(mode string) error {
	switch strings.ToLower(mode) {
	case RunServerModeAuto, RunServerModeHttp, RunServerModeSocks:
		return nil
	default:
		return fmt.Errorf("invalid server mode: %s", mode)
	}
}
