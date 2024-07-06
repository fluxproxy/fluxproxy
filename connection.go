package proxy

import (
	"io"
	"net"
)

//// Direct

var (
	_ Connection = (*DirectConnection)(nil)
)

type DirectConnection struct {
	conn net.Conn
}

func NewDirectConnection(conn net.Conn) *DirectConnection {
	return &DirectConnection{conn: conn}
}

func (d *DirectConnection) Conn() net.Conn {
	return d.conn
}

func (d *DirectConnection) Close() error {
	return d.conn.Close()
}

//// Reject

var (
	_ Connection = (*RejectConnection)(nil)
)

type RejectConnection struct {
}

func NewRejectConnection() *RejectConnection {
	return &RejectConnection{}
}

func (r *RejectConnection) Conn() net.Conn {
	return nil
}

func (r *RejectConnection) Close() error {
	return nil
}

////

type nopReadWriter struct {
}

func (nopReadWriter) Read(b []byte) (int, error) {
	return len(b), nil
}

func (nopReadWriter) Write([]byte) (int, error) {
	return 0, io.EOF
}
