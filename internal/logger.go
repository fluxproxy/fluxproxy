package internal

import (
	"context"
	"github.com/hashicorp/go-uuid"
	"github.com/sirupsen/logrus"
	"rocket/net"
	"rocket/proxy"
)

func SetupTcpContextLogger(ctx context.Context, conn *net.TCPConn) context.Context {
	id, _ := uuid.GenerateUUID()
	remoteAddr := conn.RemoteAddr()
	logger := logrus.WithFields(logrus.Fields{
		"network": "tcp",
		"address": remoteAddr.String(),
		"connid":  id,
	})
	return proxy.SetContextLogger(ctx, id, logger)
}

func SetupUdpContextLogger(ctx context.Context, conn *net.UDPAddr) context.Context {
	id, _ := uuid.GenerateUUID()
	logger := logrus.WithFields(logrus.Fields{
		"network": "udp",
		"address": conn.String(),
		"connid":  id,
	})
	return proxy.SetContextLogger(ctx, id, logger)
}
