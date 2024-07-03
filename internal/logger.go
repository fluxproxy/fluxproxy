package internal

import (
	"context"
	"github.com/hashicorp/go-uuid"
	"github.com/rocketmanapp/rocket-proxy"
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
