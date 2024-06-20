package avoidy

import (
	"avoidy/net"
	"context"
	"fmt"
	"log"
	"sync"
)

type Server struct {
	serverCtx       context.Context
	serverCtxCancel context.CancelFunc
	listeners       []Listener
	routers         []Router
	await           sync.WaitGroup
}

func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		serverCtx:       ctx,
		serverCtxCancel: cancel,
		await:           sync.WaitGroup{},
	}
}

func (s *Server) Start() error {
	// 初始化
	return nil
}

func (s *Server) Stop() error {
	s.serverCtxCancel()
	s.await.Wait()
	return nil
}

func (s *Server) Serve() error {
	if len(s.listeners) == 0 {
		return fmt.Errorf("server no any listeners")
	}
	errors := make(chan error, len(s.listeners))
	for _, listener := range s.listeners {
		s.await.Add(1)
		go func(listener Listener) {
			defer s.await.Done()
			errors <- s.serveListener(listener)
		}(listener)
	}
	select {
	case err := <-errors:
		return err
	case <-s.serverCtx.Done():
		return nil
	}
}

func (s *Server) AddListener(listener Listener) {
	s.listeners = append(s.listeners, listener)
}

func (s *Server) AddRouter(router Router) {
	s.routers = append(s.routers, router)
}

func (s *Server) serveListener(listener Listener) error {
	routers := make([]Router, len(s.routers))
	for _, router := range s.routers {
		for _, network := range router.Networks() {
			if network == listener.Network() {
				routers = append(routers, router)
			}
		}
	}
	if len(routers) == 0 {
		return fmt.Errorf("%s listener no any routers", listener.Tag())
	}
	return listener.Serve(s.serverCtx, func(ctx context.Context, conn net.Connection) {
		for _, router := range routers {
			if err := router.Route(ctx, conn); err != nil {
				log.Printf("%s listener route error. %s", listener.Tag(), err)
			}
		}
		// FIXME 多个路由？？
	})
}
