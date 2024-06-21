package net

import (
	"context"
	"io"
	"net"
)

type Connection struct {
	Context     context.Context
	Source      net.Addr // 来源地址
	Distinction net.Addr // 目标地址
	Network     Network
	Conn        *net.TCPConn
	io.ReadWriteCloser
}
