package net

import (
	"io"
	"net"
	"time"
)

func Copier(from, to net.Conn) error {
	_ = from.SetReadDeadline(time.Time{})
	_ = to.SetWriteDeadline(time.Time{})
	defer func() {
		_ = to.SetReadDeadline(time.Now()) // unlock read on 'to'
	}()
	_, err := io.Copy(to, from)
	return err
}
