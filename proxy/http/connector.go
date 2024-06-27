package http

import (
	"context"
	"crypto/tls"
	"fluxway/helper"
	"fluxway/net"
	"fluxway/proxy"
	"io"
	"net/http"
)

var (
	_ proxy.Connector = (*HrtpConnector)(nil)
)

type HrtpConnector struct {
	roundTripper http.RoundTripper
}

func NewHrtpConnector() *HrtpConnector {
	return &HrtpConnector{
		roundTripper: &http.Transport{
			TLSClientConfig: &tls.Config{},
			Proxy:           http.ProxyFromEnvironment,
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
	if resp.Body != nil {
		_, err := io.Copy(rw, resp.Body)
		return err
	} else {
		return nil
	}
}
