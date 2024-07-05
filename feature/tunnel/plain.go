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
	auth rocket.Authentication
	src  net.Address
	dest net.Address
	r    *http.Request
	w    http.ResponseWriter
	ctx  context.Context
	done context.CancelFunc
}

func NewHttpPlain(
	w http.ResponseWriter, r *http.Request, dest net.Address,
	auth rocket.Authentication,
) *HttpPlain {
	ctx, done := context.WithCancel(r.Context())
	return &HttpPlain{
		auth: auth,
		src:  auth.Source,
		dest: dest,
		r:    r,
		w:    w,
		ctx:  ctx,
		done: done,
	}
}

func (h *HttpPlain) Connect(connector rocket.Connection) {
	defer h.done()
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
	if resp.Body != nil {
		var writer io.Writer = h.w
		if len(resp.TransferEncoding) > 0 && strings.EqualFold(resp.TransferEncoding[0], "chunked") {
			writer = httpChunkWriter{Writer: h.w}
		}
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

func (h *HttpPlain) Destination() net.Address {
	return h.dest
}

func (h *HttpPlain) Source() net.Address {
	return h.auth.Source
}

func (h *HttpPlain) Authentication() rocket.Authentication {
	return h.auth
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
