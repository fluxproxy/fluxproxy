package net

import (
	"net"
	"time"
)

type TcpOptions struct {
	NoDelay      bool          `json:"no_delay"`
	KeepAlive    uint          `json:"keep_alive"`
	Linger       int           `json:"linger"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	ReadBuffer   int           `json:"read_buffer"`
	WriteBuffer  int           `json:"write_buffer"`
}

func SetTcpOptions(conn *net.TCPConn, opts TcpOptions) error {
	// No delay
	if err := conn.SetNoDelay(opts.NoDelay); err != nil {
		return err
	}
	// Keep alive
	if opts.KeepAlive > 0 {
		if err := conn.SetKeepAlive(true); err != nil {
			return err
		}
		if err := conn.SetKeepAlivePeriod(time.Duration(opts.KeepAlive) * time.Second); err != nil {
			return err
		}
	}
	// Linger
	if opts.Linger > 0 {
		if err := conn.SetLinger(opts.Linger); err != nil {
			return err
		}
	}
	// Read buffer
	if opts.ReadBuffer > 0 {
		if err := conn.SetReadBuffer(opts.ReadBuffer); err != nil {
			return err
		}
	}
	// Write buffer
	if opts.WriteBuffer > 0 {
		if err := conn.SetWriteBuffer(opts.WriteBuffer); err != nil {
			return err
		}
	}
	// Deadline defaults
	return conn.SetDeadline(time.Now().Add(time.Duration(3) * time.Second))
}

func ResetDeadline(conn *net.TCPConn, opts TcpOptions) error {
	// Read timeout
	if opts.ReadTimeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(opts.ReadTimeout)); err != nil {
			return err
		}
	}
	// Write timeout
	if opts.WriteTimeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(opts.WriteTimeout)); err != nil {
			return err
		}
	}
	return nil
}
