package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/helper"
	"github.com/sirupsen/logrus"
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
	var serverConfig ServerConfig
	if err := rocket.ConfigUnmarshalWith(runCtx, "server", &serverConfig); err != nil {
		return err
	}
	// 指定运行模式
	if serverMode == "" {
		serverConfig.Mode = RunServerModeAuto
	}
	logrus.Infof("inst: run as %s mode", serverMode)
	// 检测运行模式
	assertServerModeValid(serverConfig.Mode)
	// 启动服务端
	if helper.ContainsAnyString(serverConfig.Mode, RunServerModeForward, RunServerModeAuto) {
		if err := i.buildForwardServer(runCtx, serverConfig, serverConfig.Mode == RunServerModeForward); err != nil {
			return err
		}
	}
	if helper.ContainsAnyString(serverConfig.Mode, RunServerModeProxy, RunServerModeAuto) {
		var found = false
		// Socks server
		if ok, err := i.buildSocksServer(runCtx, serverConfig); err != nil {
			return err
		} else if ok {
			found = ok
		}
		// Http/Https server
		if ok, err := i.buildHttpServer(runCtx, serverConfig); err != nil {
			return err
		} else if ok {
			found = ok
		}
		if serverConfig.Mode == RunServerModeProxy && !found {
			return fmt.Errorf("inst: no available servers on proxy mode")
		}
	}
	// 初始化服务
	if len(i.servers) == 0 {
		return fmt.Errorf("inst: servers not found")
	}
	for _, srv := range i.servers {
		if err := srv.Init(runCtx); err != nil {
			return fmt.Errorf("inst: server init. %w", err)
		}
	}
	return nil
}

func (i *Instance) buildForwardServer(runCtx context.Context, serverConfig ServerConfig, isRequired bool) error {
	var forwardConfig ForwardConfig
	if err := rocket.ConfigUnmarshalWith(runCtx, "forward", &forwardConfig); err != nil {
		return fmt.Errorf("inst: unmarshal forward config. %w", err)
	}
	if len(forwardConfig.Rules) == 0 && isRequired {
		return fmt.Errorf("inst: empty forward rules")
	}
	for _, ruleConfig := range forwardConfig.Rules {
		if ruleConfig.Disabled {
			logrus.Warnf("inst: forward server is disabled: %s", ruleConfig.Description)
			continue
		}
		i.servers = append(i.servers, NewForwardServer(serverConfig, ruleConfig))
	}
	return nil
}

func (i *Instance) buildSocksServer(runCtx context.Context, serverConfig ServerConfig) (bool, error) {
	if serverConfig.SocksPort <= 0 {
		return false, nil
	}
	var socksConfig SocksConfig
	if err := rocket.ConfigUnmarshalWith(runCtx, "socks", &socksConfig); err != nil {
		return false, fmt.Errorf("inst: unmarshal socks config. %w", err)
	}
	if socksConfig.Disabled {
		logrus.Warnf("inst: socks server is disabled")
		return false, nil
	}
	i.servers = append(i.servers, NewSocksServer(serverConfig, socksConfig))
	return true, nil
}

func (i *Instance) buildHttpServer(runCtx context.Context, serverConfig ServerConfig) (bool, error) {
	buildServer := func(isHttps bool) error {
		var httpsConfig HttpsConfig
		if err := rocket.ConfigUnmarshalWith(runCtx, "https", &httpsConfig); err != nil {
			return fmt.Errorf("inst: unmarshal https config. %w", err)
		}
		if httpsConfig.Disabled {
			logrus.Warnf("inst: https server is disabled")
			return nil
		}
		httpsConfig.UseHttps = isHttps
		i.servers = append(i.servers, NewHttpsServer(serverConfig, httpsConfig))
		return nil
	}
	if serverConfig.HttpPort > 0 {
		if err := buildServer(false); err != nil {
			return false, err
		}
	}
	if serverConfig.HttpsPort > 0 {
		if err := buildServer(true); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (i *Instance) Serve(runCtx context.Context) error {
	servCtx, servCancel := context.WithCancel(runCtx)
	servErrors := make(chan error, len(i.servers))
	for _, srv := range i.servers {
		i.await.Add(1)
		go func(psrv rocket.Server) {
			if err := psrv.Serve(servCtx);
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
