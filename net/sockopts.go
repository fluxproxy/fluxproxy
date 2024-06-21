package net

import (
	"net"
)

func SetTcpConnOpts(conn *net.TCPConn, opts any) error {
	_ = conn.SetKeepAlive(true)
	_ = conn.SetNoDelay(true)
	_ = conn.SetReadBuffer(1024)
	_ = conn.SetWriteBuffer(1024)
	return nil
}
