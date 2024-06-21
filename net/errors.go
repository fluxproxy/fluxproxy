package net

import (
	"net"
	"vanity/common"
)

func IsTimeoutErr(err error) bool {
	nerr, ok := common.ErrorCause(err).(net.Error)
	return ok && nerr.Timeout()
}
