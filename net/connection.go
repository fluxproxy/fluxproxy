package net

import (
	"context"
	"io"
)

type Connection struct {
	// 源地址，网络类型
	Network Network // 来源地址的网络类型
	// 源地址，网络地址
	Address Address // 来源的地址
	// 源地址，Socket 读写对象
	ReadWriter io.ReadWriter
	// 源地址，Context 用于扩展
	UserContext context.Context
	// 代理目标
	Destination Destination
}

func (c Connection) WithDestination(dest Destination) Connection {
	return Connection{
		Destination: dest,
		Network:     c.Network,
		Address:     c.Address,
		ReadWriter:  c.ReadWriter,
		UserContext: c.UserContext,
	}
}
