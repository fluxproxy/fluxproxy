package helper

import (
	"fmt"
	"github.com/sirupsen/logrus"
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
