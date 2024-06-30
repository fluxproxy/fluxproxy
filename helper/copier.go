package helper

import (
	"io"
	"net"
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
	return err
}
