package net

import (
	"io"
	"net"
)

type Connection struct {
	Address     Address // 来源地址
	TCPConn     *net.TCPConn
	LongLive    bool
	Destination Destination // 目标地址
	io.ReadWriteCloser
}
