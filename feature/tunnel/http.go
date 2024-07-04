package tunnel

import (
	"context"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/helper"
	"github.com/rocket-proxy/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	"io"
	stdnet "net"
	"net/http"
	"strings"
	"time"
)

var (
	_ rocket.Tunnel = (*HttpPlain)(nil)
)

type HttpPlain struct {
	ctx  context.Context
	done context.CancelFunc
	addr net.Address
	r    *http.Request
	w    http.ResponseWriter
}

func NewHttpPlain(w http.ResponseWriter, r *http.Request, addr net.Address) *HttpPlain {
	ctx, cancel := context.WithCancel(r.Context())
	return &HttpPlain{
		addr: addr,
		r:    r,
		w:    w,
		ctx:  ctx,
		done: cancel,
	}
}

func (h *HttpPlain) Address() net.Address {
	return h.addr
}

func (h *HttpPlain) Connect(connector rocket.Connection) {
	transport := http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (stdnet.Conn, error) {
			return connector.Conn(), nil
		},
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       time.Second * 10,
		ExpectContinueTimeout: time.Second * 5,
	}
	resp, err := transport.RoundTrip(h.r)
	if err != nil {
		logrus.Errorf("Failed to connect to rocket-proxy server: %v", err)
	}
	defer helper.Close(resp.Body)

	connHeader := h.w.Header()
	for k, v := range resp.Header {
		for _, v1 := range v {
			connHeader.Add(k, v1)
		}
	}
	h.w.WriteHeader(resp.StatusCode)
	var writer io.Writer = h.w
	if len(resp.TransferEncoding) > 0 && strings.EqualFold(resp.TransferEncoding[0], "chunked") {
		writer = httpChunkWriter{Writer: h.w}
	}
	defer h.done()
	if resp.Body == nil {
		if err := helper.Copier(resp.Body, writer); err != nil {
			logrus.Errorf("Failed to copy response body: %v", err)
		}
	}
}

func (h *HttpPlain) Close() error {
	h.done()
	return nil
}

func (h *HttpPlain) Context() context.Context {
	return h.ctx
}

////

type httpChunkWriter struct {
	io.Writer
}

func (cw httpChunkWriter) Write(b []byte) (int, error) {
	n, err := cw.Writer.Write(b)
	if err == nil {
		cw.Writer.(http.Flusher).Flush()
	}
	return n, err
}
