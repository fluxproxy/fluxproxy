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
	"regexp"
	"strconv"
)

var (
	_ proxy.Listener = (*Listener)(nil)
)

var (
	regHasPort = regexp.MustCompile(`:\d+$`)
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
		logger := proxy.LoggerFromContext(r.Context())
		logger.Infof("http: proxy: %s %s", r.Method, r.URL.String())
		// Connect
		if r.Method != "CONNECT" {
			return
		}
		// Hijacker
		hij, ok := w.(http.Hijacker)
		if !ok {
			logger.Error("http not support hijack")
			return
		}
		conn, _, hijErr := hij.Hijack()
		if hijErr != nil {
			logger.Error("http not support hijack")
			return
		}
		host, port, _ := stdnet.SplitHostPort(r.URL.Host)
		connCtx, connCancel := context.WithCancel(r.Context())
		defer connCancel()
		hanErr := handler(connCtx, net.Connection{
			Network: net.Network_TCP,
			Address: net.IPAddress((conn.RemoteAddr().(*stdnet.TCPAddr)).IP),
			TCPConn: conn.(*net.TCPConn),
			Destination: net.Destination{
				Network: net.Network_TCP,
				Address: net.ParseAddress(host),
				Port:    net.ParsePort(port, 80),
			},
			ReadWriter: conn,
		})
		if hanErr == nil {
			if _, hiwErr := conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); hiwErr != nil {
				_ = conn.Close()
				if !helper.IsConnectionClosed(hiwErr) {
					logger.Errorf("http response write error: %s", hiwErr)
				}
				return
			}
		} else {
			_, _ = conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			_ = conn.Close()
			logger.Errorf("http conn error: %s", hanErr)
		}
	}
}
