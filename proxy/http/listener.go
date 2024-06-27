package http

import (
	"context"
	"fluxway/helper"
	"fluxway/internal"
	"fluxway/net"
	"fluxway/proxy"
	"github.com/sirupsen/logrus"
	stdnet "net"
	"net/http"
	"strconv"
	"strings"
)

var (
	_ proxy.Listener = (*Listener)(nil)
)

type Listener struct {
	isHttps      bool
	listenerOpts proxy.ListenerOptions
}

func NewHttpListener(isHttps bool) *Listener {
	return &Listener{
		isHttps: isHttps,
	}
}

func (l *Listener) Network() net.Network {
	return net.Network_TCP
}

func (l *Listener) ProxyType() proxy.ProxyType {
	return proxy.ProxyType_HTTPS
}

func (l *Listener) Init(options proxy.ListenerOptions) error {
	l.listenerOpts = options
	return nil
}

func (l *Listener) Serve(serveCtx context.Context, handler proxy.ListenerHandler) error {
	addr := stdnet.JoinHostPort(l.listenerOpts.Address, strconv.Itoa(l.listenerOpts.Port))
	if l.isHttps {
		logrus.Infof("http: serve start, https, address: %s", addr)
	} else {
		logrus.Infof("http: serve start, address: %s", addr)
	}
	server := &http.Server{
		Addr:    addr,
		Handler: l.newServeHandler(handler),
		BaseContext: func(l stdnet.Listener) context.Context {
			return serveCtx
		},
		ConnContext: func(ctx context.Context, conn stdnet.Conn) context.Context {
			return internal.SetupTcpContextLogger(serveCtx, conn.(*net.TCPConn))
		},
	}
	defer func() {
		logrus.Infof("http serve stop, address: %s", addr)
		_ = server.Shutdown(serveCtx)
	}()
	if l.isHttps {
		return server.ListenAndServeTLS(l.listenerOpts.TLSCertFile, l.listenerOpts.TLSKeyFile)
	} else {
		return server.ListenAndServe()
	}
}

func (l *Listener) newServeHandler(handler proxy.ListenerHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := proxy.RequiredLogger(r.Context())
		logger.Infof("http: %s %s", r.Method, r.RequestURI)
		// Auth: nop
		// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Proxy-Authenticate
		// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Proxy-Authorization
		r.Header.Del("Proxy-Connection")
		r.Header.Del("Proxy-Authenticate")
		r.Header.Del("Proxy-Authorization")
		// Connect
		if r.Method != "CONNECT" {
			return
		}
		// Hijacker
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			logger.Error("http: not support hijack")
			return
		}
		hijConn, _, hijErr := hijacker.Hijack()
		if hijErr != nil {
			logger.Error("http: not support hijack")
			return
		}
		defer net.Close(hijConn)
		connCtx, connCancel := context.WithCancel(r.Context())
		defer connCancel()
		// Phase hook
		connCtx = proxy.ContextWithHookDialPhased(connCtx, func(ctx context.Context, conn *net.Connection) error {
			if _, hiwErr := conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); hiwErr != nil {
				if !helper.IsConnectionClosed(hiwErr) {
					logger.Errorf("http: write back ok response: %s", hiwErr)
				}
				return hiwErr
			}
			return nil
		})
		addr, port, _ := parseHostToAddress(r.URL.Host)
		hErr := handler(connCtx, net.Connection{
			Network: net.Network_TCP,
			Address: net.IPAddress((hijConn.RemoteAddr().(*stdnet.TCPAddr)).IP),
			TCPConn: hijConn.(*net.TCPConn),
			Destination: net.Destination{
				Network: net.Network_TCP,
				Address: addr,
				Port:    port,
			},
			ReadWriter: hijConn,
		})
		if hErr != nil {
			_, _ = hijConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
			logger.Errorf("http: conn handle: %s", hErr)
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
