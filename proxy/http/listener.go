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
	options proxy.ListenerOptions
}

func NewHttpListener() *Listener {
	return &Listener{}
}

func (l *Listener) Network() net.Network {
	return net.Network_TCP
}

func (l *Listener) ProxyType() proxy.ProxyType {
	return proxy.ProxyType_HTTPS
}

func (l *Listener) Init(options proxy.ListenerOptions) error {
	l.options = options
	return nil
}

func (l *Listener) Serve(serveCtx context.Context, handler proxy.ListenerHandler) error {
	addr := stdnet.JoinHostPort(l.options.Address, strconv.Itoa(l.options.Port))
	logrus.Infof("http: serve start, address: %s", addr)
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
	return server.ListenAndServe()
}

func (l *Listener) newServeHandler(handler proxy.ListenerHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := proxy.RequiredLogger(r.Context())
		logger.Infof("http: proxy: %s %s", r.Method, r.URL.String())
		// Auth: nop
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
			logger.Error("http not support hijack")
			return
		}
		conn, _, hijErr := hijacker.Hijack()
		if hijErr != nil {
			logger.Error("http not support hijack")
			return
		}
		connCtx, connCancel := context.WithCancel(r.Context())
		defer connCancel()
		// Phase hook
		connCtx = proxy.ContextWithHookDialPhased(connCtx, func(ctx context.Context, conn *net.Connection) error {
			if _, hiwErr := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); hiwErr != nil {
				if !helper.IsConnectionClosed(hiwErr) {
					logger.Errorf("http response write error: %s", hiwErr)
				}
				return hiwErr
			}
			return nil
		})
		addr, port, _ := parseHostToAddress(r.URL.Host)
		hanErr := handler(connCtx, net.Connection{
			Network: net.Network_TCP,
			Address: net.IPAddress((conn.RemoteAddr().(*stdnet.TCPAddr)).IP),
			TCPConn: conn.(*net.TCPConn),
			Destination: net.Destination{
				Network: net.Network_TCP,
				Address: addr,
				Port:    port,
			},
			ReadWriter: conn,
		})
		if hanErr != nil {
			_, _ = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			_ = conn.Close()
			logger.Errorf("http conn error: %s", hanErr)
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
