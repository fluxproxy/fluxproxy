package internal

import (
	"context"
	"github.com/fluxproxy/fluxproxy"
	"strings"
)

func LogTailError(connCtx context.Context, tag string, disErr error) {
	msg := disErr.Error()
	if strings.Contains(msg, "i/o timeout") {
		return
	}
	if strings.Contains(msg, "connection reset by peer") {
		return
	}
	proxy.Logger(connCtx).Errorf("%s conn error: %s", tag, disErr)
}
