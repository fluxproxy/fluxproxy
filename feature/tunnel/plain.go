package tunnel

import (
	"context"
	"fmt"
	"github.com/fluxproxy/fluxproxy/helper"
	"github.com/fluxproxy/fluxproxy/net"
	"io"
	stdnet "net"
	"net/http"
	"strings"
	"time"
)

var (
	_ proxy.Tunnel = (*HttpPlainTunnel)(nil)
)

type HttpPlainTunnel struct {
	src        net.Address
	dest       net.Address
	r          *http.Request
	w          http.ResponseWriter
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewHttpPlain(
	w http.ResponseWriter,
	r *http.Request,
	dest net.Address,
	src net.Address,
) *HttpPlainTunnel {
	ctx, cancel := context.WithCancel(r.Context())
	return &HttpPlainTunnel{
		src:        src,
		dest:       dest,
		r:          r,
		w:          w,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (h *HttpPlainTunnel) Connect(connection proxy.Connection) error {
	defer h.cancelFunc()
	transport := http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (stdnet.Conn, error) {
			return connection.Conn(), nil
		},
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       time.Second * 10,
		ExpectContinueTimeout: time.Second * 5,
	}
	resp, rtErr := transport.RoundTrip(h.r)
	if rtErr != nil {
		return fmt.Errorf("http: roundtrip. %w", rtErr)
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
		return helper.Copier(resp.Body, writer)
	}
	return nil
}

func (h *HttpPlainTunnel) Close() error {
	h.cancelFunc()
	return nil
}

func (h *HttpPlainTunnel) Context() context.Context {
	return h.ctx
}

func (h *HttpPlainTunnel) Destination() net.Address {
	return h.dest
}

func (h *HttpPlainTunnel) Source() net.Address {
	return h.src
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
