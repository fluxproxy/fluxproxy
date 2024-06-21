package common

import (
	"fmt"
	"io"
	"net"
	"time"
)

func Copy(reader, writer net.Conn, errors chan<- error) {
	_ = reader.SetReadDeadline(time.Time{})
	_ = writer.SetWriteDeadline(time.Time{})
	if _, err := io.Copy(writer, reader); err == nil {
		errors <- nil // A successful copy end
	} else {
		errors <- fmt.Errorf("remote-conn end")
	}
}
