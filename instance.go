package fluxway

import (
	"context"
	"fluxway/helper"
	"fluxway/proxy"
	"fmt"
	"github.com/bytepowered/assert-go"
	"github.com/sirupsen/logrus"
	"sync"
)

type Instance struct {
	instCtx       context.Context
	instCtxCancel context.CancelFunc
	servers       []proxy.Server
	await         sync.WaitGroup
}

func NewInstance(runCtx context.Context) *Instance {
	if runCtx == nil {
		runCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(runCtx)
	return &Instance{
		instCtx:       ctx,
		instCtxCancel: cancel,
		await:         sync.WaitGroup{},
	}
}

func (i *Instance) Start(startAsMode string) error {
	// 解析配置
	var serverOpts ServerOptions
	if err := proxy.UnmarshalConfig(i.instCtx, "server", &serverOpts); err != nil {
		return err
	}
	// 指定运行模式
	if startAsMode != "" {
		serverOpts.Mode = startAsMode
		logrus.Warnf("inst: run as server mode: %s", startAsMode)
	}
	// 检测运行模式
	AssertServerModeValid(serverOpts.Mode)
	// 启动服务端
	if helper.ContainsAnyString(serverOpts.Mode, ServerModeForward, ServerModeMixin) {
		if err := i.buildForwardServer(serverOpts); err != nil {
			return err
		}
	}
	if helper.ContainsAnyString(serverOpts.Mode, ServerModeProxy, ServerModeMixin) {
		if err := i.buildProxyServer(serverOpts); err != nil {
			return err
		}
	}
	// 初始化服务
	assert.MustTrue(len(i.servers) > 0, "servers is required")
	for _, server := range i.servers {
		if err := server.Init(i.instCtx); err != nil {
			return fmt.Errorf("server init error. %w", err)
		}
	}
	return nil
}

func (i *Instance) buildForwardServer(serverOpts ServerOptions) error {
	var forwardOpts ForwardRootOptions
	if err := proxy.UnmarshalConfig(i.instCtx, "forward", &forwardOpts); err != nil {
		return fmt.Errorf("unmarshal forward options: %w", err)
	}
	assert.MustTrue(len(forwardOpts.Rules) > 0, "forward options is required")
	for _, rule := range forwardOpts.Rules {
		if rule.Disabled {
			logrus.Warnf("inst: forward server is disabled: %s", rule.Description)
			continue
		}
		i.servers = append(i.servers, NewForwardServer(serverOpts, rule))
	}
	return nil
}

func (i *Instance) buildProxyServer(serverOpts ServerOptions) error {
	// Socks proxy
	var socksOpts SocksOptions
	if err := proxy.UnmarshalConfig(i.instCtx, "socks", &socksOpts); err != nil {
		return fmt.Errorf("unmarshal socks options: %w", err)
	}
	if socksOpts.Disabled {
		logrus.Warnf("inst: socks server is disabled")
		return nil
	}
	i.servers = append(i.servers, NewSocksServer(serverOpts, socksOpts))
	return nil
}

func (i *Instance) Stop() error {
	i.instCtxCancel()
	i.await.Wait()
	return nil
}

func (i *Instance) Serve() error {
	logrus.Info("inst: serve start")
	defer logrus.Info("inst: serve stop")
	if len(i.servers) == 0 {
		return fmt.Errorf("servers is required")
	}
	errors := make(chan error, len(i.servers))
	for _, server := range i.servers {
		i.await.Add(1)
		go func(server proxy.Server, ctx context.Context) {
			defer i.await.Done()
			errors <- server.Serve(ctx)
		}(server, i.instCtx)
	}
	select {
	case err := <-errors:
		return err
	case <-i.instCtx.Done():
		return nil
	}
}
