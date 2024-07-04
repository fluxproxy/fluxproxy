package connector

import (
	"github.com/rocket-proxy/rocket-proxy"
	"io"
	"net"
)

var (
	_ rocket.Connector = (*Direct)(nil)
)

type Direct struct {
	conn net.Conn
}

func NewDirect(conn net.Conn) *Direct {
	return &Direct{conn: conn}
}

func (d *Direct) ReadWriter() io.ReadWriter {
	return d.conn
}

func (d *Direct) Conn() net.Conn {
	return d.conn
}

func (d *Direct) Close() error {
	return d.conn.Close()
}
