package net

import (
	"context"
	"io"
	"net"
)

type Connection struct {
	// 目标
	Destination Destination
	// 来源
	Network Network // 来源地址的网络类型
	Address Address // 来源的地址
	// 来源 Socket: tcp
	TCPConn *net.TCPConn // TCP连接（仅当 Network 为 TCP类型时）
	// 来源 Socket 读写对象
	ReadWriter io.ReadWriter
	// 来源 Context 用于扩展
	UserContext context.Context
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
