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
	return rocket.SetContextLogID(ctx, id, remoteAddr.String())
}

func SetupUdpContextLogger(ctx context.Context, conn *net.UDPAddr) context.Context {
	id, _ := uuid.GenerateUUID()
	return rocket.SetContextLogID(ctx, id, conn.String())
}
