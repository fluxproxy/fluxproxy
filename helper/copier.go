package helper

import (
	"io"
	"net"
	"strings"
	"time"
)

func Copier(from, to io.ReadWriter) error {
	fromConn, fromConnOK := from.(net.Conn)
	if fromConnOK {
		_ = fromConn.SetReadDeadline(time.Time{})
	}
	toConn, toConnOK := to.(net.Conn)
	if toConnOK {
		_ = toConn.SetWriteDeadline(time.Time{})
		defer func() {
			_ = toConn.SetReadDeadline(time.Now()) // unlock read on 'to'
		}()
	}
	_, err := io.Copy(to, from)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "use of closed network connection") {
			return io.EOF
		}
		if strings.Contains(msg, "i/o timeout") {
			return io.EOF
		}
		if strings.Contains(msg, "connection reset by peer") {
			return io.EOF
		}
		return err
	}
	return nil
}
