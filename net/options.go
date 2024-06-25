package net

import (
	"net"
	"time"
)

type TcpOptions struct {
	NoDelay      bool          `json:"no_delay"`
	KeepAlive    time.Duration `json:"keep_alive"`
	Linger       int           `json:"linger"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	ReadBuffer   int           `json:"read_buffer"`
	WriteBuffer  int           `json:"write_buffer"`
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

func SetTcpOptions(conn net.Conn, opts TcpOptions) error {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		// No delay
		if err := tcpConn.SetNoDelay(opts.NoDelay); err != nil {
			return err
		}
		// Keep alive
		if opts.KeepAlive > 0 {
			if err := tcpConn.SetKeepAlive(true); err != nil {
				return err
			}
			if err := tcpConn.SetKeepAlivePeriod(opts.KeepAlive); err != nil {
				return err
			}
		}
		// Linger
		if opts.Linger > 0 {
			if err := tcpConn.SetLinger(opts.Linger); err != nil {
				return err
			}
		}
		// Read buffer
		if opts.ReadBuffer > 0 {
			if err := tcpConn.SetReadBuffer(opts.ReadBuffer); err != nil {
				return err
			}
		}
		// Write buffer
		if opts.WriteBuffer > 0 {
			if err := tcpConn.SetWriteBuffer(opts.WriteBuffer); err != nil {
				return err
			}
		}
	}
	// Deadline defaults
	return conn.SetDeadline(time.Time{})
}

//// UDP

type UdpOptions struct {
}

func DefaultUdpOptions() UdpOptions {
	return UdpOptions{}
}
