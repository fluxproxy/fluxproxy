package helper

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
	"syscall"
)

func ErrIf(err error, message string) error {
	if err != nil {
		return fmt.Errorf(message, err)
	}
	return nil
}

func LogIf(err error, message string) {
	if err != nil {
		logrus.Errorf(message, err)
	}
}

func IsConnectionClosed(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	if errors.Is(err, context.Canceled) {
		return true
	}
	i := 0
	var newErr = &err
	for opError, ok := (*newErr).(*net.OpError); ok && i < 10; {
		i++
		newErr = &opError.Err
		if syscallError, ok := (*newErr).(*os.SyscallError); ok {
			if syscallError.Err == syscall.EPIPE || syscallError.Err == syscall.ECONNRESET || syscallError.Err == syscall.EPROTOTYPE {
				return true
			}
		}
	}
	return false
}
