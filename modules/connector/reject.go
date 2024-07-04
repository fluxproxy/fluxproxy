package connector

import (
	"github.com/rocket-proxy/rocket-proxy"
	"io"
	"net"
)

var (
	_ rocket.Connector = (*Reject)(nil)
)

type Reject struct {
}

func NewReject() *Reject {
	return &Reject{}
}

func (r *Reject) ReadWriter() io.ReadWriter {
	return &NopRw{}
}

func (r *Reject) Conn() net.Conn {
	return nil
}

func (r *Reject) Close() error {
	return nil
}

////

type NopRw struct{}

func (NopRw) Read(b []byte) (int, error) {
	return len(b), nil
}

func (NopRw) Write([]byte) (int, error) {
	return 0, io.EOF
}
