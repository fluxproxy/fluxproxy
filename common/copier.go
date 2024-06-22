package common

import (
	"fmt"
	"github.com/valyala/bytebufferpool"
	"io"
	"net"
	"time"
)

func Copy(from, to net.Conn, errors chan<- error) {
	_ = from.SetReadDeadline(time.Time{})
	_ = to.SetWriteDeadline(time.Time{})
	buffer := bytebufferpool.Get()
	buffer.Reset()
	defer bytebufferpool.Put(buffer)
	if _, err := io.Copy(to, from); err == nil {
		errors <- nil // A successful copy end
	} else {
		errors <- fmt.Errorf("remote-conn end")
	}
}
