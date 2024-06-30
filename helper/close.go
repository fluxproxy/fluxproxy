package helper

import (
	"io"
	"net"
	"time"
)

func Close(v any) {
	if v == nil {
		return
	}
	if c, ok := v.(io.Closer); ok {
		closeCloser(c)
	} else if n, ok := v.(net.Conn); ok {
		closeConn(n)
	}
}

func closeCloser(conn io.Closer) {
	_ = conn.Close()
}

func closeConn(conn net.Conn) {
	_ = conn.SetDeadline(time.Now().Add(-time.Second))
	_ = conn.Close()
}
