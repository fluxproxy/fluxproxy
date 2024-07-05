package listener

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytepowered/goes"
	stdnet "net"
	"time"
)

func tcpListenWith(serveCtx context.Context, opts proxy.ListenerOptions, connHandler func(*stdnet.TCPConn)) error {
	addr := &stdnet.TCPAddr{IP: stdnet.ParseIP(opts.Address), Port: opts.Port}
	listener, lErr := stdnet.ListenTCP("tcp", addr)
	if lErr != nil {
		return fmt.Errorf("listen %s. %w", addr, lErr)
	}
	_ = listener.SetDeadline(time.Time{})
	go func() {
		<-serveCtx.Done()
		_ = listener.Close()
	}()
	var tempDelay time.Duration
	for {
		conn, acErr := listener.Accept()
		if acErr != nil {
			select {
			case <-serveCtx.Done():
				return serveCtx.Err()
			default:
				var netErr stdnet.Error
				if errors.As(acErr, &netErr) && netErr.Temporary() {
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if maxDuration := 1 * time.Second; tempDelay > maxDuration {
						tempDelay = maxDuration
					}
					time.Sleep(tempDelay)
					continue
				}
				return fmt.Errorf("accept. %w", acErr)
			}
		}
		goes.Go(func() {
			connHandler(conn.(*stdnet.TCPConn))
		})
	}
}
