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
	if _, err := io.Copy(to, from); err == nil {
		return nil // A successful copy end
	} else {
		return err
	}
}
