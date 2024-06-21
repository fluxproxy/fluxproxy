package net

import (
	"net"
	"time"
)

var (
	ZeroDuration = time.Unix(0, 0)
)

type TcpOptions struct {
	NoDelay      bool          `json:"no_delay"`
	KeepAlive    time.Duration `json:"keep_alive"`
	Linger       int           `json:"linger"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	ReadBuffer   int           `json:"read_buffer"`
	WriteBuffer  int           `json:"write_buffer"`
	AwaitTimeout time.Duration `json:"await_timeout"`
}

func DefaultTcpOptions() TcpOptions {
	return TcpOptions{
		ReadTimeout:  0,
		WriteTimeout: 0,
		ReadBuffer:   1024,
		WriteBuffer:  1024,
		NoDelay:      true,
		KeepAlive:    time.Second * 10,
	}
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
		if err := conn.SetKeepAlivePeriod(opts.KeepAlive); err != nil {
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
	return conn.SetDeadline(time.Time{})
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

func ResetReadDeadline(conn *net.TCPConn, opts TcpOptions) error {
	// Read timeout
	if opts.ReadTimeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(opts.ReadTimeout)); err != nil {
			return err
		}
	}
	return nil
}

func ResetWriteDeadline(conn *net.TCPConn, opts TcpOptions) error {
	// Write timeout
	if opts.WriteTimeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(opts.WriteTimeout)); err != nil {
			return err
		}
	}
	return nil
}

func Close(conn net.Conn) {
	_ = conn.SetDeadline(time.Now().Add(-time.Second))
	_ = conn.Close()
}
