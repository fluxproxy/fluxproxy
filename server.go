package vanity

import (
	"context"
	"fmt"
	"sync"
	"vanity/common"
	"vanity/net"
)

type Server struct {
	serverCtx       context.Context
	serverCtxCancel context.CancelFunc
	listeners       []Listener
	dispatchers     []*Dispatcher
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
		newDispatcher := NewDispatcher(listener, s.findRouters(listener.Network()))
		if len(newDispatcher.routers) == 0 {
			return fmt.Errorf("%s listener no any routers", listener.Tag())
		}
		s.dispatchers = append(s.dispatchers, newDispatcher)
		go func(listener Listener, dispatcher *Dispatcher) {
			defer s.await.Done()
			ctx := toContext(s.serverCtx, s)
			errors <- listener.Serve(ctx, dispatcher.Process)
		}(listener, newDispatcher)
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

func (s *Server) findRouters(network net.Network) []Router {
	common.Assert(network != net.Network_Unknown, "network must not be unknown")
	routers := make([]Router, len(s.routers))
	for _, router := range s.routers {
		for _, n := range router.Networks() {
			if n != net.Network_Unknown && n == network {
				routers = append(routers, router)
			}
		}
	}
	return routers
}
