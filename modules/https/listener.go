package https

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"github.com/bytepowered/assert"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/helper"
	"github.com/rocketmanapp/rocket-proxy/internal"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	stdnet "net"
	"net/http"
	"strconv"
	"strings"
)

var (
	_ rocket.Listener = (*Listener)(nil)
)

type Options struct {
	UseHttps bool
}

type Listener struct {
	opts         Options
	listenerOpts rocket.ListenerOptions
	roundTripper http.RoundTripper
}

func NewHttpsListener(opts Options) *Listener {
	return &Listener{
		opts: opts,
		roundTripper: &http.Transport{
			TLSClientConfig: &tls.Config{},
			Proxy:           http.ProxyFromEnvironment,
		},
	}
}

func (l *Listener) Network() net.Network {
	return net.Network_TCP
}

func (l *Listener) Init(options rocket.ListenerOptions) error {
	l.listenerOpts = options
	return nil
}

func (l *Listener) Listen(serveCtx context.Context, dispatchHandler rocket.ListenerHandler) error {
	addr := stdnet.JoinHostPort(l.listenerOpts.Address, strconv.Itoa(l.listenerOpts.Port))
	if l.opts.UseHttps {
		logrus.Infof("https: listen start, HTTPS, address: %s", addr)
	} else {
		logrus.Infof("https: listen start, address: %s", addr)
	}
	server := &http.Server{
		Addr:    addr,
		Handler: l.newServeHandler(dispatchHandler),
		BaseContext: func(l stdnet.Listener) context.Context {
			return serveCtx
		},
		ConnContext: func(ctx context.Context, conn stdnet.Conn) context.Context {
			return internal.SetupTcpContextLogger(ctx, conn.(*net.TCPConn))
		},
	}
	go func() {
		<-serveCtx.Done()
		_ = server.Shutdown(serveCtx)
	}()
	if l.opts.UseHttps {
		return server.ListenAndServeTLS(l.listenerOpts.TLSCertFile, l.listenerOpts.TLSKeyFile)
	} else {
		return server.ListenAndServe()
	}
}

func (l *Listener) newServeHandler(dispatchHandler rocket.ListenerHandler) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rocket.Logger(r.Context()).Infof("https: %s %s", r.Method, r.RequestURI)
		if r.Method == http.MethodConnect {
			l.handleConnectStream(rw, r, dispatchHandler)
		} else {
			l.handlePlainRequest(rw, r, dispatchHandler)
		}
	}
}

func (l *Listener) handleConnectStream(rw http.ResponseWriter, r *http.Request, dispatchHandler rocket.ListenerHandler) {
	connCtx, connCancel := context.WithCancel(r.Context())
	defer connCancel()
	// Hijacker
	r = r.WithContext(connCtx)
	hijacker, ok := rw.(http.Hijacker)
	assert.MustTrue(ok, "https: not support hijack")
	hijConn, _, hijErr := hijacker.Hijack()
	if hijErr != nil {
		_, _ = rw.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		rocket.Logger(connCtx).Error("https: not support hijack")
		return
	}
	defer helper.Close(hijConn)
	addr, port, _ := parseHostToAddress(r.URL.Host)
	conn := net.Connection{
		Network:     l.Network(),
		Address:     net.IPAddress((hijConn.RemoteAddr().(*stdnet.TCPAddr)).IP),
		ReadWriter:  hijConn.(*net.TCPConn),
		UserContext: setWithUserContext(context.Background(), rw, r),
		Destination: net.Destination{
			Network: net.Network_TCP,
			Address: addr,
			Port:    port,
		},
	}
	// Auth
	aErr := dispatchHandler.Auth(connCtx, conn, parseProxyAuthorization(r.Header))
	if aErr != nil {
		_, _ = hijConn.Write([]byte("HTTP/1.1 401 Unauthorized\r\n\r\n"))
		rocket.Logger(connCtx).Errorf("https: conn auth: %s", aErr)
		return
	} else {
		removeHopByHopHeaders(r.Header)
	}
	// Phase hook
	connCtx = rocket.ContextWithHookFuncDialPhased(connCtx, func(ctx context.Context, conn *net.Connection) error {
		if _, hiwErr := hijConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); hiwErr != nil {
			if !helper.IsConnectionClosed(hiwErr) {
				rocket.Logger(connCtx).Errorf("https: write back ok response: %s", hiwErr)
			}
			return hiwErr
		}
		return nil
	})
	// Next
	hErr := dispatchHandler.Handle(connCtx, conn)
	// Complete
	if hErr != nil {
		_, _ = hijConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		rocket.Logger(connCtx).Errorf("https: conn handle: %s", hErr)
	}
}

func (l *Listener) handlePlainRequest(rw http.ResponseWriter, r *http.Request, dispatchHandler rocket.ListenerHandler) {
	defer helper.Close(r.Body)

	if r.URL.Host == "" || !r.URL.IsAbs() {
		// RFC 2068 (HTTP/1.1) requires URL to be absolute URL in HTTP proxy.
		response := &http.Response{
			Status:        "Bad Request",
			StatusCode:    400,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        http.Header(make(map[string][]string)),
			Body:          nil,
			ContentLength: 0,
			Close:         true,
		}
		response.Header.Set("Proxy-Connection", "close")
		response.Header.Set("Connection", "close")
		_ = response.Write(rw)
		return
	}
	// Forward host
	if len(r.URL.Host) > 0 {
		r.Host = r.URL.Host
	}
	// Prevent UA from being set to golang's default ones
	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", "")
	}
	// Next
	connCtx := r.Context()
	addr, port, _ := parseHostToAddress(r.URL.Host)
	conn := net.Connection{
		Network:     l.Network(),
		Address:     net.ParseAddress(r.RemoteAddr),
		UserContext: setWithUserContext(context.Background(), rw, r),
		ReadWriter:  nil,
		Destination: net.Destination{
			Network: net.Network_HRTP,
			Address: addr,
			Port:    port,
		},
	}
	// Auth
	aErr := dispatchHandler.Auth(connCtx, conn, parseProxyAuthorization(r.Header))
	if aErr != nil {
		_, _ = rw.Write([]byte("HTTP/1.1 401 Unauthorized\r\n\r\n"))
		rocket.Logger(connCtx).Errorf("https: conn auth: %s", aErr)
		return
	} else {
		removeHopByHopHeaders(r.Header)
	}
	// Next
	hErr := dispatchHandler.Handle(connCtx, conn)
	// Complete
	if hErr != nil {
		_, _ = rw.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		rocket.Logger(connCtx).Errorf("https: conn handle: %s", hErr)
	}
}

func parseProxyAuthorization(header http.Header) rocket.ListenerAuthorization {
	authorization := header.Get("Proxy-Authorization")
	if strings.HasPrefix(authorization, rocket.AuthenticateBasic) {
		username, password, _ := parseBasicAuth(authorization)
		return rocket.ListenerAuthorization{
			Authenticate:  rocket.AuthenticateBasic,
			Authorization: username + ":" + password,
		}
	} else if strings.HasPrefix(authorization, rocket.AuthenticateBearer) {
		token, _ := parseBearerAuth(authorization)
		return rocket.ListenerAuthorization{
			Authenticate:  rocket.AuthenticateBearer,
			Authorization: token,
		}
	} else {
		return rocket.ListenerAuthorization{
			Authenticate:  rocket.AuthenticateToken,
			Authorization: authorization,
		}
	}
}

func parseHostToAddress(urlHost string) (addr net.Address, port net.Port, err error) {
	if strings.LastIndexByte(urlHost, ':') > 0 {
		h, p, e := stdnet.SplitHostPort(urlHost)
		if e != nil {
			return nil, 0, e
		}
		addr = net.ParseAddress(h)
		port = net.ParsePort(p, 80)
	} else {
		addr = net.ParseAddress(urlHost)
		port = net.Port(80)
	}
	return addr, port, nil
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !helper.ASCIIEqualFold(auth[:len(prefix)], prefix) {
		return "", "", false
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
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

func parseBearerAuth(auth string) (token string, ok bool) {
	const prefix = "Bearer "
	if len(auth) < len(prefix) || !helper.ASCIIEqualFold(auth[:len(prefix)], prefix) {
		return "", false
	}
	return auth[len(prefix):], true
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
