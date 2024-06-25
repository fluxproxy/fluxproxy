package net

import (
	"io"
	"net"
	"time"
)

var ()

func Copier(from, to net.Conn) error {
	_ = from.SetReadDeadline(time.Time{})
	_ = to.SetWriteDeadline(time.Time{})
	defer func() {
		_ = to.SetReadDeadline(time.Now()) // unlock read on 'to'
	}()
	buffer := make([]byte, 1024)
	if _, err := io.CopyBuffer(to, from, buffer); err == nil {
		return nil // A successful copy end
	} else {
		return io.EOF
	}
}

func Close(conn net.Conn) {
	_ = conn.SetDeadline(time.Now().Add(-time.Second))
	_ = conn.Close()
}
