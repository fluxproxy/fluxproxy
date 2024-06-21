package net

import (
	"io"
	"net"
)

type Connection struct {
	Address Address // 来源地址
	TCPConn *net.TCPConn
	io.ReadWriteCloser
}

type Link struct {
	Connection  *Connection
	KeepAlive   bool
	Destination Destination // 目标地址
}
