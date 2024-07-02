package https

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/helper"
	"github.com/rocketmanapp/rocket-proxy/net"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	_ rocket.Connector = (*HrtpConnector)(nil)
)

type HrtpConnector struct {
	roundTripper http.RoundTripper
}

func NewHrtpConnector() *HrtpConnector {
	return &HrtpConnector{
		roundTripper: &http.Transport{
			// from https.DefaultTransport
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func (c *HrtpConnector) DialServe(connCtx context.Context, link *net.Connection) error {
	w := requiredResponseWriter(link.UserContext)
	r := requiredHttpRequest(link.UserContext)
	return c.httpRoundTrip(connCtx, w, r)
}

func (c *HrtpConnector) httpRoundTrip(ctx context.Context, rw http.ResponseWriter, r *http.Request) error {
	resp, err := c.roundTripper.RoundTrip(r)
	if err != nil {
		return err
	} else {
		defer helper.Close(resp.Body)
	}
	connHeader := rw.Header()
	for k, v := range resp.Header {
		for _, v1 := range v {
			connHeader.Add(k, v1)
		}
	}
	rw.WriteHeader(resp.StatusCode)
	var writer io.Writer = rw
	if len(resp.TransferEncoding) > 0 && strings.EqualFold(resp.TransferEncoding[0], "chunked") {
		writer = httpChunkWriter{Writer: rw}
	}
	if resp.Body != nil {
		_, err := io.Copy(writer, resp.Body)
		return err
	} else {
		return nil
	}
}

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
