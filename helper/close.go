package helper

import (
	"io"
	"net"
	"time"
)

func Close(c io.Closer) {
	if c == nil {
		return
	}
	if conn, ok := c.(net.Conn); ok {
		closeConn(conn)
	} else {
		closeCloser(c)
	}
}

func closeCloser(conn io.Closer) {
	_ = conn.Close()
}

func closeConn(conn net.Conn) {
	_ = conn.SetDeadline(time.Now().Add(-time.Second))
	_ = conn.Close()
}
