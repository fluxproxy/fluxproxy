package internal

import (
	"context"
	"fluxway/net"
	"fluxway/proxy"
	"github.com/hashicorp/go-uuid"
	"github.com/sirupsen/logrus"
)

func SetupTcpContextLogger(ctx context.Context, conn *net.TCPConn) context.Context {
	id, _ := uuid.GenerateUUID()
	remoteAddr := conn.RemoteAddr()
	logger := logrus.WithFields(logrus.Fields{
		"network": "tcp",
		"address": remoteAddr.String(),
		"connid":  id,
	})
	ctx = proxy.ContextWithID(ctx, id)
	ctx = proxy.ContextWithLogger(ctx, logger)
	return ctx
}

func SetupUdpContextLogger(ctx context.Context, conn *net.UDPAddr) context.Context {
	id, _ := uuid.GenerateUUID()
	logger := logrus.WithFields(logrus.Fields{
		"network": "udp",
		"address": conn.String(),
		"connid":  id,
	})
	ctx = proxy.ContextWithID(ctx, id)
	ctx = proxy.ContextWithLogger(ctx, logger)
	return ctx
}
