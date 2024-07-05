package internal

import (
	"context"
	"github.com/lithammer/shortuuid/v4"
	"github.com/rocket-proxy/rocket-proxy"
	"net"
	"time"
)

var (
	CtxKeyStartTime = "ctx-key:start-time"
)

func SetupTcpContextLogger(ctx context.Context, conn net.Conn) context.Context {
	id := shortuuid.New()
	remoteAddr := conn.RemoteAddr()
	return setContextLogID(ctx, id, remoteAddr.String())
}

func SetupUdpContextLogger(ctx context.Context, conn *net.UDPAddr) context.Context {
	id := shortuuid.New()
	remoteAddr := conn.String()
	return setContextLogID(ctx, id, remoteAddr)
}

func setContextLogID(ctx context.Context, id string, source string) context.Context {
	return context.WithValue(context.WithValue(context.WithValue(ctx,
		rocket.CtxKeyID, id),
		CtxKeyStartTime, time.Now()),
		rocket.CtxKeySource, source)
}
