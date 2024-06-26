package net

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

type Port uint16

func PortFromBytes(port []byte) Port {
	return Port(binary.BigEndian.Uint16(port))
}

func PortFromInt(val uint32) (Port, error) {
	if val > 65535 {
		return Port(0), fmt.Errorf("invalid port range: %d", val)
	}
	return Port(val), nil
}

func PortFromString(s string) (Port, error) {
	val, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return Port(0), fmt.Errorf("invalid port range: %s", s)
	}
	return PortFromInt(uint32(val))
}

func (p Port) Value() uint16 {
	return uint16(p)
}

func (p Port) String() string {
	return strconv.Itoa(int(p))
}
