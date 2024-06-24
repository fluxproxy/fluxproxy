package net

import (
	"fmt"
	"io"
	"net"
	"time"
)

func Copier(from, to net.Conn) error {
	_ = from.SetReadDeadline(time.Time{})
	_ = to.SetWriteDeadline(time.Time{})
	defer to.SetReadDeadline(time.Now()) // unlock read on 'to'
	if _, err := io.Copy(to, from); err == nil {
		return nil // A successful copy end
	} else {
		return fmt.Errorf("remote-conn end")
	}
}

func Close(conn net.Conn) {
	_ = conn.SetDeadline(time.Now().Add(-time.Second))
	_ = conn.Close()
}
