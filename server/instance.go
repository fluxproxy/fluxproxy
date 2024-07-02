package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/helper"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"sync"
)

const (
	RunServerModeAuto    string = "auto"
	RunServerModeProxy   string = "proxy"
	RunServerModeForward string = "forward"
)

type Instance struct {
	servers []rocket.Server
	await   sync.WaitGroup
}

func NewInstance() *Instance {
	return &Instance{
		await: sync.WaitGroup{},
	}
}

func (i *Instance) Init(runCtx context.Context, serverMode string) error {
	// 解析配置
	var serverOpts Options
	if err := rocket.ConfigUnmarshalWith(runCtx, "server", &serverOpts); err != nil {
		return err
	}
	// 指定运行模式
	if serverMode == "" {
		serverOpts.Mode = RunServerModeAuto
	}
	logrus.Info("inst: run as server mode: ", serverMode)
	// 检测运行模式
	assertServerModeValid(serverOpts.Mode)
	// 启动服务端
	if helper.ContainsAnyString(serverOpts.Mode, RunServerModeForward, RunServerModeAuto) {
		if err := i.buildForwardServer(runCtx, serverOpts, serverOpts.Mode == RunServerModeForward); err != nil {
			return err
		}
	}
	if helper.ContainsAnyString(serverOpts.Mode, RunServerModeProxy, RunServerModeAuto) {
		var found = false
		// Socks server
		if ok, err := i.buildSocksServer(runCtx, serverOpts); err != nil {
			return err
		} else if ok {
			found = ok
		}
		// Http/Https server
		if ok, err := i.buildHttpServer(runCtx, serverOpts); err != nil {
			return err
		} else if ok {
			found = ok
		}
		if serverOpts.Mode == RunServerModeProxy && !found {
			return fmt.Errorf("proxy servers not found")
		}
	}
	// 初始化服务
	if len(i.servers) == 0 {
		return fmt.Errorf("servers not found")
	}
	for _, srv := range i.servers {
		if err := srv.Init(runCtx); err != nil {
			return fmt.Errorf("server init error. %w", err)
		}
	}
	return nil
}

func (i *Instance) buildForwardServer(runCtx context.Context, serverOpts Options, isRequired bool) error {
	var forwardOpts ForwardRootOptions
	if err := rocket.ConfigUnmarshalWith(runCtx, "forward", &forwardOpts); err != nil {
		return fmt.Errorf("unmarshal forward options: %w", err)
	}
	if len(forwardOpts.Rules) == 0 && isRequired {
		return fmt.Errorf("forward rules is empty")
	}
	for _, rule := range forwardOpts.Rules {
		if rule.Disabled {
			logrus.Warnf("inst: forward server is disabled: %s", rule.Description)
			continue
		}
		i.servers = append(i.servers, NewForwardServer(serverOpts, rule))
	}
	return nil
}

func (i *Instance) buildSocksServer(runCtx context.Context, serverOpts Options) (bool, error) {
	if serverOpts.SocksPort <= 0 {
		return false, nil
	}
	var socksOpts SocksOptions
	if err := rocket.ConfigUnmarshalWith(runCtx, "socks", &socksOpts); err != nil {
		return false, fmt.Errorf("unmarshal socks options: %w", err)
	}
	if socksOpts.Disabled {
		logrus.Warnf("inst: socks server is disabled")
		return false, nil
	}
	i.servers = append(i.servers, NewSocksServer(serverOpts, socksOpts))
	return true, nil
}

func (i *Instance) buildHttpServer(runCtx context.Context, serverOpts Options) (bool, error) {
	buildServer := func(serverOpts Options, isHttps bool) error {
		var httpOpts HttpsOptions
		if err := rocket.ConfigUnmarshalWith(runCtx, "https", &httpOpts); err != nil {
			return fmt.Errorf("unmarshal https options: %w", err)
		}
		if httpOpts.Disabled {
			logrus.Warnf("inst: https server is disabled")
			return nil
		}
		i.servers = append(i.servers, NewHttpsServer(serverOpts, httpOpts, isHttps))
		return nil
	}
	if serverOpts.HttpPort > 0 {
		if err := buildServer(serverOpts, false); err != nil {
			return false, err
		}
	}
	if serverOpts.HttpsPort > 0 {
		if err := buildServer(serverOpts, true); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (i *Instance) Serve(runCtx context.Context) error {
	if len(i.servers) == 0 {
		return fmt.Errorf("servers is required")
	}
	servErrors := make(chan error, len(i.servers))
	for _, srv := range i.servers {
		i.await.Add(1)
		go func(psrv rocket.Server) {
			if err := psrv.Serve(runCtx);
				err == nil ||
					errors.Is(err, io.EOF) ||
					errors.Is(err, context.Canceled) ||
					errors.Is(err, http.ErrServerClosed) ||
					helper.IsConnectionClosed(err) {
				servErrors <- nil
			} else {
				servErrors <- err
			}
			i.await.Done()
		}(srv)
	}
	select {
	case err := <-servErrors:
		return i.term(err)
	case <-runCtx.Done():
		return i.term(nil)
	}
}

func (i *Instance) term(err error) error {
	i.await.Wait()
	return err
}

func assertServerModeValid(mode string) {
	valid := false
	switch strings.ToLower(mode) {
	case RunServerModeForward, RunServerModeAuto, RunServerModeProxy:
		valid = true
	default:
		valid = false
	}
	assert.MustTrue(valid, "server mode is invalid, was: %s", mode)
}
