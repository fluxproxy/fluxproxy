package listener

import (
	"context"
	"encoding/base64"
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
	// Hijacker
	hijacker, ok := rw.(http.Hijacker)
	assert.MustTrue(ok, "http: not support hijack")
	hiConn, _, hijErr := hijacker.Hijack()
	if hijErr != nil {
		_, _ = rw.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		rocket.Logger(r.Context()).Error("http: not support hijack")
		return
	}
	defer helper.Close(hiConn)

	srcAddr := parseRemoteToSrcAddress(r.RemoteAddr)
	destAddr := parseHostToDestAddress(r.Host)
	auth := parseProxyAuthorization(r.Header, srcAddr)

	removeHopByHopHeaders(r.Header)

	connCtx := internal.ContextWithHook(r.Context(), internal.CtxHookAfterDialed, newConnEstablishedHook(hiConn))
	connCtx = internal.ContextWithHook(connCtx, internal.CtxHookAfterAuthed, newUnauthorizedHook(hiConn))

	stream := tunnel.NewConnStream(connCtx, hiConn, destAddr, auth)
	defer helper.Close(stream)
	dispatcher.Submit(stream)

	<-stream.Context().Done()
}

func (l *HttpListener) handlePlainRequest(rw http.ResponseWriter, r *http.Request, dispatcher rocket.Dispatcher) {
	// RFC 2068 (HTTP/1.1) requires URL to be absolute URL in HTTP proxy.
	if r.URL.Host == "" || !r.URL.IsAbs() {
		_, _ = rw.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}
	defer helper.Close(r.Body)

	// Prevent UA from being set to golang's default ones
	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", "")
	}

	srcAddr := parseRemoteToSrcAddress(r.RemoteAddr)
	destAddr := parseHostToDestAddress(r.Host)
	auth := parseProxyAuthorization(r.Header, srcAddr)

	removeHopByHopHeaders(r.Header)

	connCtx := internal.ContextWithHook(r.Context(), internal.CtxHookAfterDialed, newConnEstablishedHook(rw))
	connCtx = internal.ContextWithHook(connCtx, internal.CtxHookAfterAuthed, newUnauthorizedHook(rw))

	plain := tunnel.NewHttpPlain(rw, r.WithContext(connCtx), destAddr, auth)
	defer helper.Close(plain)
	dispatcher.Submit(plain)

	<-plain.Context().Done()
}

func newUnauthorizedHook(w io.Writer) rocket.HookFunc {
	return func(ctx context.Context, state error, _ ...any) error {
		if state == nil {
			return nil
		}
		rocket.Logger(ctx).Errorf("http: conn auth: %s", state)
		_, err := w.Write([]byte("HTTP/1.1 401 Unauthorized\r\n\r\n"))
		if err != nil {
			return fmt.Errorf("http send response. %w", err)
		}
		return nil
	}
}

func newConnEstablishedHook(w io.Writer) rocket.HookFunc {
	return func(_ context.Context, _ error, _ ...any) error {
		_, err := w.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		if err != nil {
			return fmt.Errorf("http send response. %w", err)
		}
		return nil
	}
}

func removeHopByHopHeaders(header http.Header) {
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

func parseProxyAuthorization(header http.Header, srcAddr net.Address) rocket.Authentication {
	token := header.Get("Proxy-Authorization")
	if strings.HasPrefix(token, rocket.AuthenticateBasic) {
		username, password, _ := parseBasicAuthorization(token)
		return rocket.Authentication{
			Source:         srcAddr,
			Authenticate:   rocket.AuthenticateBasic,
			Authentication: username + ":" + password,
		}
	} else if strings.HasPrefix(token, rocket.AuthenticateBearer) {
		token, _ := parseBearerAuthorization(token)
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

func parseBasicAuthorization(token string) (username, password string, ok bool) {
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

func parseBearerAuthorization(token string) (out string, ok bool) {
	const prefix = "Bearer "
	if len(token) < len(prefix) || !helper.ASCIIEqualFold(token[:len(prefix)], prefix) {
		return "", false
	}
	return token[len(prefix):], true
}

func parseHostToDestAddress(host string) (addr net.Address) {
	assert.MustNotEmpty(host, "http host is empty")
	if strings.LastIndexByte(host, ':') > 0 {
		addr, _ = net.ParseAddress(net.NetworkTCP, host)
	} else {
		addr, _ = net.ParseAddress(net.NetworkTCP, host+":80")
	}
	return
}

func parseRemoteToSrcAddress(remoteAddr string) net.Address {
	host, _, hpErr := stdnet.SplitHostPort(remoteAddr)
	assert.MustNil(hpErr, "http: parse host port error: %s", hpErr)
	srcAddr, _ := net.ParseAddress(net.NetworkTCP, host)
	assert.MustTrue(srcAddr.IsIP(), "http: srcAddr is not ip")
	return srcAddr
}
