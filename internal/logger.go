package internal

import (
	"context"
	"github.com/hashicorp/go-uuid"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/helper"
	"github.com/rocketmanapp/rocket-proxy/net"
)

func SetupTcpContextLogger(ctx context.Context, conn *net.TCPConn) context.Context {
	id, _ := uuid.GenerateUUID()
	remoteAddr := conn.RemoteAddr()
	return rocket.SetContextLogID(ctx, id, remoteAddr.String())
}

func SetupUdpContextLogger(ctx context.Context, conn *net.UDPAddr) context.Context {
	id, _ := uuid.GenerateUUID()
	return rocket.SetContextLogID(ctx, id, conn.String())
}

func onTailError(connCtx context.Context, tag string, disErr error) {
	if disErr == nil {
		return
	}
	if !helper.IsCopierError(disErr) {
		rocket.Logger(connCtx).Errorf("%s conn error: %s", tag, disErr)
	}
}
