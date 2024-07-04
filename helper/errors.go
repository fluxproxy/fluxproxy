package helper

import (
	"fmt"
)

func ErrIf(err error, message string) error {
	if err != nil {
		return fmt.Errorf(message, err)
	}
	return nil
}
