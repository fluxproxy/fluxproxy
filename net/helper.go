package net

import (
	"fmt"
	"io"
	"net"
	"time"
)

func Copied(from, to net.Conn, errors chan<- error) {
	_ = from.SetReadDeadline(time.Time{})
	_ = to.SetWriteDeadline(time.Time{})
	if _, err := io.Copy(to, from); err == nil {
		errors <- nil // A successful copy end
	} else {
		errors <- fmt.Errorf("remote-conn end")
	}
}

func Close(conn net.Conn) {
	_ = conn.SetDeadline(time.Now().Add(-time.Second))
	_ = conn.Close()
}
