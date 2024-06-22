package common

import (
    "fmt"
    "github.com/valyala/bytebufferpool"
    "io"
    "net"
    "time"
)

func Copy(reader, writer net.Conn, errors chan<- error) {
    _ = reader.SetReadDeadline(time.Time{})
    _ = writer.SetWriteDeadline(time.Time{})
    buffer := bytebufferpool.Get()
    defer bytebufferpool.Put(buffer)
    if _, err := io.CopyBuffer(writer, reader, buffer.B); err == nil {
        errors <- nil // A successful copy end
    } else {
        errors <- fmt.Errorf("remote-conn end")
    }
}
