package helper

import (
	"errors"
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
	if err == nil {
		return io.EOF
	}
	return io.ErrUnexpectedEOF
}

func IsCopierError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	return errors.Is(err, io.ErrUnexpectedEOF)
}
