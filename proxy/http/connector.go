package http

import (
	"context"
	"crypto/tls"
	"errors"
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

func (c *HrtpConnector) DialServe(inctx context.Context, link *net.Connection) error {
	w := requiredResponseWriter(inctx)
	r := requiredHttpRequest(inctx)
	return c.httpRoundTrip(inctx, w, r)
}

func (c *HrtpConnector) httpRoundTrip(ctx context.Context, connWriter http.ResponseWriter, r *http.Request) error {
	resp, err := c.roundTripper.RoundTrip(r)
	if err != nil {
		if !errors.Is(err, context.Canceled) && helper.IsConnectionClosed(err) {
			proxy.Logger(ctx).Error("http: forward round trip: %s", err)
		}
		return err
	} else {
		defer helper.Close(resp.Body)
	}
	connHeader := connWriter.Header()
	for k, v := range resp.Header {
		for _, v1 := range v {
			connHeader.Add(k, v1)
		}
	}
	connWriter.WriteHeader(resp.StatusCode)
	if resp.Body != nil {
		_, err := io.Copy(connWriter, resp.Body)
		return err
	} else {
		return nil
	}
}
