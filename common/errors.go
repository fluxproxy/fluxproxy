package common

import "os"

// ErrorCause returns the root cause of this error.
func ErrorCause(err error) error {
	if err == nil {
		return nil
	}
L:
	for {
		switch inner := err.(type) {
		case *os.PathError:
			if inner.Err == nil {
				break L
			}
			err = inner.Err
		case *os.SyscallError:
			if inner.Err == nil {
				break L
			}
			err = inner.Err
		default:
			break L
		}
	}
	return err
}
