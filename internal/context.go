package internal

import (
	"context"
	"github.com/hashicorp/go-uuid"
	"github.com/rocket-proxy/rocket-proxy"
	"net"
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
	ctx = context.WithValue(ctx, rocket.CtxKeyID, id)
	return context.WithValue(ctx, rocket.CtxKeySource, source)
}
