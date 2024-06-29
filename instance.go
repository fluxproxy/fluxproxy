package fluxway

import (
	"context"
	"errors"
	"fluxway/helper"
	"fluxway/proxy"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"sync"
)

type Instance struct {
	//instCtx       context.Context
	//instCtxCancel context.CancelFunc
	servers []proxy.Server
	await   sync.WaitGroup
}

func NewInstance() *Instance {
	return &Instance{
		await: sync.WaitGroup{},
	}
}

func (i *Instance) Init(runCtx context.Context, startAsMode string) error {
	// 解析配置
	var serverOpts ServerOptions
	if err := proxy.UnmarshalConfig(runCtx, "server", &serverOpts); err != nil {
		return err
	}
	// 指定运行模式
	if startAsMode != "" {
		serverOpts.Mode = startAsMode
	}
	logrus.Info("inst: run as server mode: ", startAsMode)
	// 检测运行模式
	assertServerModeValid(serverOpts.Mode)
	// 启动服务端
	if helper.ContainsAnyString(serverOpts.Mode, ServerModeForward, ServerModeMixin) {
		if err := i.buildForwardServer(runCtx, serverOpts, serverOpts.Mode == ServerModeForward); err != nil {
			return err
		}
	}
	if helper.ContainsAnyString(serverOpts.Mode, ServerModeProxy, ServerModeMixin) {
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
		if serverOpts.Mode == ServerModeProxy && !found {
			return fmt.Errorf("proxy servers not found")
		}
	}
	// 初始化服务
	if len(i.servers) == 0 {
		return fmt.Errorf("servers not found")
	}
	for _, server := range i.servers {
		if err := server.Init(runCtx); err != nil {
			return fmt.Errorf("server init error. %w", err)
		}
	}
	return nil
}

func (i *Instance) buildForwardServer(runCtx context.Context, serverOpts ServerOptions, isRequired bool) error {
	var forwardOpts ForwardRootOptions
	if err := proxy.UnmarshalConfig(runCtx, "forward", &forwardOpts); err != nil {
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

func (i *Instance) buildSocksServer(runCtx context.Context, serverOpts ServerOptions) (bool, error) {
	if serverOpts.SocksPort <= 0 {
		return false, nil
	}
	var socksOpts SocksOptions
	if err := proxy.UnmarshalConfig(runCtx, "socks", &socksOpts); err != nil {
		return false, fmt.Errorf("unmarshal socks options: %w", err)
	}
	if socksOpts.Disabled {
		logrus.Warnf("inst: socks server is disabled")
		return false, nil
	}
	i.servers = append(i.servers, NewSocksServer(serverOpts, socksOpts))
	return true, nil
}

func (i *Instance) buildHttpServer(runCtx context.Context, serverOpts ServerOptions) (bool, error) {
	buildServer := func(serverOpts ServerOptions, isHttps bool) error {
		var httpOpts HttpOptions
		if err := proxy.UnmarshalConfig(runCtx, "http", &httpOpts); err != nil {
			return fmt.Errorf("unmarshal http options: %w", err)
		}
		if httpOpts.Disabled {
			logrus.Warnf("inst: http server is disabled")
			return nil
		}
		i.servers = append(i.servers, NewHttpServer(serverOpts, httpOpts, isHttps))
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
	for _, server := range i.servers {
		i.await.Add(1)
		go func(server proxy.Server) {
			if err := server.Serve(runCtx);
				errors.Is(err, io.EOF) ||
					errors.Is(err, context.Canceled) ||
					errors.Is(err, http.ErrServerClosed) {
				servErrors <- nil
			} else {
				servErrors <- err
			}
			i.await.Done()
		}(server)
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
