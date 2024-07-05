package internal

import (
	"context"
	"github.com/hashicorp/go-uuid"
	"github.com/rocket-proxy/rocket-proxy"
	"net"
	"time"
)

var (
	CtxKeyStartTime = "ctx-key:start-time"
)

func SetupTcpContextLogger(ctx context.Context, conn net.Conn) context.Context {
	id, _ := uuid.GenerateUUID()
	remoteAddr := conn.RemoteAddr()
	return setContextLogID(ctx, id, remoteAddr.String())
}

func SetupUdpContextLogger(ctx context.Context, conn *net.UDPAddr) context.Context {
	id, _ := uuid.GenerateUUID()
	return setContextLogID(ctx, id, conn.String())
}

func setContextLogID(ctx context.Context, id string, source string) context.Context {
	return context.WithValue(context.WithValue(context.WithValue(ctx,
		rocket.CtxKeyID, id),
		CtxKeyStartTime, time.Now()),
		rocket.CtxKeySource, source)
}
