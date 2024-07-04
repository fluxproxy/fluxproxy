package listener

import (
	"context"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/feature/tunnel"
	"github.com/rocket-proxy/rocket-proxy/helper"
	"github.com/rocket-proxy/rocket-proxy/internal"
	"github.com/rocket-proxy/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	stdnet "net"
	"net/http"
	"strconv"
	"strings"
)

var (
	_ rocket.Listener = (*HttpListener)(nil)
)

type HttpListener struct {
	listenerOpts rocket.ListenerOptions
}

func NewHttpListener(opts rocket.ListenerOptions) *HttpListener {
	return &HttpListener{
		listenerOpts: opts,
	}
}

func (l *HttpListener) Init(ctx context.Context) error {
	if l.listenerOpts.Port <= 0 {
		return fmt.Errorf("http: invalid port: %d", l.listenerOpts.Port)
	}
	return nil
}

func (l *HttpListener) Listen(serveCtx context.Context, dispatcher rocket.Dispatcher) error {
	addr := stdnet.JoinHostPort(l.listenerOpts.Address, strconv.Itoa(l.listenerOpts.Port))
	logrus.Infof("http: listen: %s", addr)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: l.newServeHandler(dispatcher),
		BaseContext: func(_ stdnet.Listener) context.Context {
			return serveCtx
		},
		ConnContext: func(ctx context.Context, conn stdnet.Conn) context.Context {
			return internal.SetupTcpContextLogger(ctx, conn)
		},
	}
	go func() {
		<-serveCtx.Done()
		_ = httpServer.Shutdown(serveCtx)
	}()
	return httpServer.ListenAndServe()
}

func (l *HttpListener) newServeHandler(dispatcher rocket.Dispatcher) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		logrus.Infof("http: %s %s", r.Method, r.RequestURI)
		if r.Method == http.MethodConnect {
			l.handleConnectStream(rw, r, dispatcher)
		} else {
			l.handlePlainRequest(rw, r, dispatcher)
		}
	}
}

func (l *HttpListener) handleConnectStream(rw http.ResponseWriter, r *http.Request, dispatcher rocket.Dispatcher) {
	connCtx := r.Context()
	// Hijacker
	hijacker, ok := rw.(http.Hijacker)
	assert.MustTrue(ok, "http: not support hijack")
	hijConn, _, hijErr := hijacker.Hijack()
	if hijErr != nil {
		_, _ = rw.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		rocket.Logger(connCtx).Error("http: not support hijack")
		return
	}
	// Authenticate

	//hook: on dialer
	connCtx = rocket.ContextWithHookFunc(connCtx, rocket.CtxHookFuncOnDialer, func(context.Context) error {
		_, err := hijConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		if err != nil {
			return fmt.Errorf("http send response. %w", err)
		}
		return nil
	})

	stream := tunnel.NewHttpStream(connCtx, hijConn, parseHostToAddress(r.URL.Host))
	defer helper.Close(stream)
	dispatcher.Submit(stream)

	<-stream.Done()
}

func (l *HttpListener) handlePlainRequest(rw http.ResponseWriter, r *http.Request, dispatcher rocket.Dispatcher) {
	// RFC 2068 (HTTP/1.1) requires URL to be absolute URL in HTTP proxy.
	if r.URL.Host == "" || !r.URL.IsAbs() {
		_, _ = rw.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	// Prevent UA from being set to golang's default ones
	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", "")
	}

	// Authenticate

	mono := tunnel.NewHttpPlain(rw, r, parseHostToAddress(r.Host))
	defer helper.Close(mono)
	dispatcher.Submit(mono)

	<-mono.Done()
}

func parseHostToAddress(host string) (addr net.Address) {
	assert.MustNotEmpty(host, "http host is empty")
	if strings.LastIndexByte(host, ':') > 0 {
		addr, _ = net.ParseAddress(net.NetworkTCP, host)
	} else {
		addr, _ = net.ParseAddress(net.NetworkTCP, host+":80")
	}
	return
}
