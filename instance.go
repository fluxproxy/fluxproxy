package fluxway

import (
	"context"
	"fluxway/common"
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

func (i *Instance) Start() error {
	logrus.Infof("instance start")
	//// 解析配置
	k := proxy.ConfigFromContext(i.instCtx)
	var serverOpts ServerOptions
	if err := k.Unmarshal("server", &serverOpts); err != nil {
		return fmt.Errorf("unmarshal server options: %w", err)
	}
	if common.StringContains(serverOpts.Mode, "forward", "mixin") {
		if err := i.buildForwardServer(serverOpts); err != nil {
			return err
		}
	}
	if common.StringContains(serverOpts.Mode, "proxy", "mixin") {
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
	k := proxy.ConfigFromContext(i.instCtx)
	type RootForwarderOptions []ForwardOptions
	var opts RootForwarderOptions
	if err := k.Unmarshal("forward", &opts); err != nil {
		return fmt.Errorf("unmarshal forward options: %w", err)
	}
	assert.MustTrue(len(opts) > 0, "forward options is required")
	for _, opt := range opts {
		if opt.Disabled {
			logrus.Warnf("forward server is disabled: %s", opt.Description)
			continue
		}
		i.servers = append(i.servers, NewForwardServer(serverOpts, opt))
	}
	return nil
}

func (i *Instance) buildProxyServer(serverOpts ServerOptions) error {
	//k := proxy.ConfigFromContext(i.instCtx)
	return nil
}

func (i *Instance) Stop() error {
	i.instCtxCancel()
	i.await.Wait()
	logrus.Infof("instance stop")
	return nil
}

func (i *Instance) Serve() error {
	logrus.Infof("instance serve")
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
