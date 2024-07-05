package listener

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/feature/tunnel"
	"github.com/rocket-proxy/rocket-proxy/helper"
	"github.com/rocket-proxy/rocket-proxy/internal"
	"github.com/rocket-proxy/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	"io"
	stdnet "net"
	"net/http"
	"strconv"
	"strings"
)

var (
	_ rocket.Listener = (*HttpListener)(nil)
)

type HttpOptions struct {
}

type HttpListener struct {
	opts         HttpOptions
	listenerOpts rocket.ListenerOptions
}

func NewHttpListener(listenerOpts rocket.ListenerOptions, httpOpts HttpOptions) *HttpListener {
	return &HttpListener{
		listenerOpts: listenerOpts,
		opts:         httpOpts,
	}
}

func (l *HttpListener) Init(runCtx context.Context) error {
	if l.listenerOpts.Port <= 0 {
		return fmt.Errorf("http: invalid port: %d", l.listenerOpts.Port)
	}
	return nil
}

func (l *HttpListener) Listen(serveCtx context.Context, dispatcher rocket.Dispatcher) error {
	addr := stdnet.JoinHostPort(l.listenerOpts.Address, strconv.Itoa(l.listenerOpts.Port))
	if l.listenerOpts.Auth {
		logrus.Infof("http: listen: %s", addr)
	} else {
		logrus.Infof("http: listen(no auth): %s", addr)
	}
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
		if r.Method == http.MethodConnect {
			l.handleConnectStream(rw, r, dispatcher)
		} else {
			l.handlePlainRequest(rw, r, dispatcher)
		}
	}
}

func (l *HttpListener) handleConnectStream(rw http.ResponseWriter, r *http.Request, dispatcher rocket.Dispatcher) {
	// Hijacker
	hijacker, ok := rw.(http.Hijacker)
	assert.MustTrue(ok, "http: not support hijack")
	hiConn, _, hiErr := hijacker.Hijack()
	if hiErr != nil {
		rw.WriteHeader(http.StatusBadGateway)
		rocket.Logger(r.Context()).Error("http: not support hijack")
		return
	}

	srcAddr := parseRemoteAddress(r.RemoteAddr)
	destAddr := l.parseHostAddress(r.Host)

	if l.listenerOpts.Auth {
		auErr := dispatcher.Authenticate(r.Context(), l.parseProxyAuthorization(r.Header, srcAddr))
		if auErr != nil {
			_, _ = hiConn.Write([]byte("HTTP/1.1 401 Unauthorized\r\n\r\n"))
			return
		}
	}
	l.removeHopByHopHeaders(r.Header)

	ctx := internal.ContextWithHooks(r.Context(), map[any]rocket.HookFunc{
		internal.CtxHookAfterRuleset: l.withRulesetHook(hiConn),
		internal.CtxHookAfterDialed:  l.withDialedHook(hiConn, r),
	})

	stream := tunnel.NewConnStream(ctx, hiConn, destAddr, srcAddr)
	dispatcher.Submit(stream)

	if l.listenerOpts.Verbose {
		rocket.Logger(r.Context()).
			WithField("dest", r.Host).
			WithField("rdest", destAddr.String()).
			Infof("http: %s", r.Method)
	}
}

func (l *HttpListener) handlePlainRequest(rw http.ResponseWriter, r *http.Request, dispatcher rocket.Dispatcher) {
	// RFC 2068 (HTTP/1.1) requires URL to be absolute URL in HTTP proxy.
	if r.URL.Host == "" || !r.URL.IsAbs() {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", "")
	}
	srcAddr := parseRemoteAddress(r.RemoteAddr)

	// Authenticate
	if l.listenerOpts.Auth {
		auErr := dispatcher.Authenticate(r.Context(), l.parseProxyAuthorization(r.Header, srcAddr))
		if auErr != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
	}
	l.removeHopByHopHeaders(r.Header)

	// Destination
	destAddr := l.parseHostAddress(r.Host)

	// Submit
	ctx := internal.ContextWithHooks(r.Context(), map[any]rocket.HookFunc{
		internal.CtxHookAfterRuleset: l.withRulesetHook(rw),
		internal.CtxHookAfterDialed:  l.withDialedHook(rw, r),
	})
	plain := tunnel.NewHttpPlain(rw, r.WithContext(ctx), destAddr, srcAddr)
	dispatcher.Submit(plain)

	if l.listenerOpts.Verbose {
		rocket.Logger(r.Context()).
			WithField("dest", r.Host).
			WithField("rdest", destAddr.String()).
			Infof("http: %s", r.Method)
	}
}

func (*HttpListener) withRulesetHook(w io.Writer) rocket.HookFunc {
	return func(ctx context.Context, state error, _ ...any) error {
		if state == nil || errors.Is(state, rocket.ErrNoRulesetMatched) {
			return nil
		}
		rocket.Logger(ctx).Errorf("http: conn ruleset: %s", state)
		if rw, ok := w.(http.ResponseWriter); ok {
			rw.WriteHeader(http.StatusForbidden)
		} else {
			_, err := w.Write([]byte("HTTP/1.1 403 Forbidden\r\n\r\n"))
			if err != nil {
				return fmt.Errorf("http send response(ruleset). %w", err)
			}
		}
		return errors.New("ruleset forbidden")
	}
}

func (*HttpListener) withDialedHook(w io.Writer, r *http.Request) rocket.HookFunc {
	return func(_ context.Context, _ error, _ ...any) error {
		if rw, ok := w.(http.ResponseWriter); ok {
			rw.WriteHeader(http.StatusOK)
		} else {
			_, err := w.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
			if err != nil {
				return fmt.Errorf("http send response(established). %w", err)
			}
		}
		return nil
	}
}

func (*HttpListener) removeHopByHopHeaders(header http.Header) {
	// Strip hop-by-hop header based on RFC:
	// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html#sec13.5.1
	// https://www.mnot.net/blog/2011/07/11/what_proxies_must_do

	header.Del("Proxy-Connection")
	header.Del("Proxy-Authenticate")
	header.Del("Proxy-Authorization")
	header.Del("TE")
	header.Del("Trailers")
	header.Del("Transfer-Encoding")
	header.Del("Upgrade")
	header.Del("Keep-Alive")

	connections := header.Get("Connection")
	header.Del("Connection")
	if connections == "" {
		return
	}
	for _, h := range strings.Split(connections, ",") {
		header.Del(strings.TrimSpace(h))
	}
}

func (l *HttpListener) parseProxyAuthorization(header http.Header, srcAddr net.Address) rocket.Authentication {
	token := header.Get("Proxy-Authorization")
	if strings.HasPrefix(token, "Basic ") {
		username, password, _ := l.parseBasicAuthorization(token)
		return rocket.Authentication{
			Source:         srcAddr,
			Authenticate:   rocket.AuthenticateBasic,
			Authentication: username + ":" + password,
		}
	} else if strings.HasPrefix(token, "Bearer ") {
		token, _ := l.parseBearerAuthorization(token)
		return rocket.Authentication{
			Source:         srcAddr,
			Authenticate:   rocket.AuthenticateBearer,
			Authentication: token,
		}
	} else {
		return rocket.Authentication{
			Source:         srcAddr,
			Authenticate:   rocket.AuthenticateToken,
			Authentication: token,
		}
	}
}

func (*HttpListener) parseBasicAuthorization(token string) (username, password string, ok bool) {
	const prefix = "Basic "
	if len(token) < len(prefix) || !helper.ASCIIEqualFold(token[:len(prefix)], prefix) {
		return "", "", false
	}
	c, err := base64.StdEncoding.DecodeString(token[len(prefix):])
	if err != nil {
		return "", "", false
	}
	cs := string(c)
	username, password, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false
	}
	return username, password, true
}

func (*HttpListener) parseBearerAuthorization(token string) (out string, ok bool) {
	const prefix = "Bearer "
	if len(token) < len(prefix) || !helper.ASCIIEqualFold(token[:len(prefix)], prefix) {
		return "", false
	}
	return token[len(prefix):], true
}

func (*HttpListener) parseHostAddress(host string) (addr net.Address) {
	assert.MustNotEmpty(host, "http host is empty")
	if strings.LastIndexByte(host, ':') > 0 {
		addr, _ = net.ParseAddress(net.NetworkTCP, host)
	} else {
		addr, _ = net.ParseAddress(net.NetworkTCP, host+":80")
	}
	return
}
