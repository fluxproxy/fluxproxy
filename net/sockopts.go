package net

import (
	"net"
	"time"
)

func SetTcpConnOpts(conn *net.TCPConn, opts any) error {
	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
	_ = conn.SetKeepAlive(true)
	_ = conn.SetNoDelay(true)
	_ = conn.SetReadBuffer(1024)
	_ = conn.SetWriteBuffer(1024)
	return nil
}
