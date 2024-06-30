package net

import (
	"net"
	"time"
)

type TcpOptions struct {
	NoDelay     bool          `json:"no_delay"`
	KeepAlive   time.Duration `json:"keep_alive"`
	Linger      int           `json:"linger"`
	ReadBuffer  int           `json:"read_buffer"`
	WriteBuffer int           `json:"write_buffer"`
}

func DefaultTcpOptions() TcpOptions {
	return TcpOptions{
		ReadBuffer:  1024 * 32,
		WriteBuffer: 1024 * 32,
		NoDelay:     true,
		KeepAlive:   time.Second * 10,
	}
}

func SetTcpConnOptions(conn net.Conn, opts TcpOptions) error {
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
	ReadBuffer  int `json:"read_buffer"`
	WriteBuffer int `json:"write_buffer"`
}

func DefaultUdpOptions() UdpOptions {
	return UdpOptions{
		ReadBuffer:  1024 * 32,
		WriteBuffer: 1024 * 32,
	}
}

func SetUdpConnOptions(conn net.Conn, opts UdpOptions) error {
	if udpConn, ok := conn.(*net.UDPConn); ok {
		// Read buffer
		if opts.ReadBuffer > 0 {
			if err := udpConn.SetReadBuffer(opts.ReadBuffer); err != nil {
				return err
			}
		}
		// Write buffer
		if opts.WriteBuffer > 0 {
			if err := udpConn.SetWriteBuffer(opts.WriteBuffer); err != nil {
				return err
			}
		}
	}
	// Deadline defaults
	return conn.SetDeadline(time.Time{})
}
