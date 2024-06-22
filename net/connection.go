package net

import (
	"io"
	"net"
)

type Connection struct {
	// 目标
	Destination Destination
	// 来源
	Network Network      // 网络类型
	Address Address      // 地址
	TCPConn *net.TCPConn // TCP连接（仅当 Network 为 TCP类型时）
	io.ReadWriter
}

func (c Connection) WithDestination(dest Destination) Connection {
	return Connection{
		Destination: dest,
		Network:     c.Network,
		Address:     c.Address,
		TCPConn:     c.TCPConn,
		ReadWriter:  c.ReadWriter,
	}
}
