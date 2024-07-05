package net

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type AddressFamily byte

const (
	AddressFamilyIPv4   = AddressFamily(0)
	AddressFamilyIPv6   = AddressFamily(1)
	AddressFamilyDomain = AddressFamily(2)
)

type Address struct {
	Network Network
	Family  AddressFamily
	IP      net.IP
	Domain  string
	Port    int
}

func (a Address) Addrport() string {
	sport := strconv.Itoa(a.Port)
	switch a.Family {
	case AddressFamilyDomain:
		return a.Domain + ":" + sport
	case AddressFamilyIPv6:
		return "[" + a.IP.String() + "]:" + sport
	case AddressFamilyIPv4:
		fallthrough
	default:
		return a.IP.String() + ":" + sport
	}
}

func (a Address) Addr() string {
	switch a.Family {
	case AddressFamilyDomain:
		return a.Domain
	case AddressFamilyIPv6:
		return "[" + a.IP.String() + "]"
	case AddressFamilyIPv4:
		fallthrough
	default:
		return a.IP.String()
	}
}

func (a Address) String() string {
	return a.Network.String() + "://" + a.Addrport()
}

func (a Address) IsIP() bool {
	return a.Family != AddressFamilyDomain
}

func (a Address) IsDomain() bool {
	return a.Family == AddressFamilyDomain
}

func (a Address) Equal(o Address) bool {
	if a.Family != o.Family {
		return false
	}
	if a.Port != o.Port {
		return false
	}
	if a.Family == AddressFamilyDomain {
		return a.Domain == o.Domain
	} else {
		return a.IP.Equal(o.IP)
	}
}

func ToAddressFamily(ip net.IP) AddressFamily {
	if ip.To4() != nil {
		return AddressFamilyIPv4
	}
	return AddressFamilyIPv6
}

func ParseDomainAddr(network Network, domain string) Address {
	return Address{
		Network: network,
		Family:  AddressFamilyDomain,
		Domain:  domain,
	}
}

func ParseIPAddr(network Network, ipAddr net.IP) Address {
	if ipAddr.To4() != nil {
		return Address{
			Network: network,
			Family:  AddressFamilyIPv4,
			IP:      ipAddr,
		}
	}
	return Address{
		Network: network,
		Family:  AddressFamilyIPv6,
		IP:      ipAddr,
	}
}

func ParseAddress(network Network, hostport string) (Address, error) {
	addr, sport, err := net.SplitHostPort(hostport)
	if err != nil {
		return Address{}, fmt.Errorf("invalid host:port address: %w", err)
	}
	port, err := strconv.Atoi(sport)
	if err != nil {
		return Address{}, fmt.Errorf("invalid port number: %w", err)
	}
	// Handle IPv6 address in form as "[2001:4860:0:2001::68]"
	lenAddr := len(addr)
	if lenAddr > 0 && addr[0] == '[' && addr[lenAddr-1] == ']' {
		addr = addr[1 : lenAddr-1]
		lenAddr -= 2
	}
	if lenAddr > 0 && (!isAlphaNum(addr[0]) || !isAlphaNum(addr[len(addr)-1])) {
		addr = strings.TrimSpace(addr)
	}
	ip := net.ParseIP(addr)
	if ip == nil {
		return Address{
			Network: network,
			Family:  AddressFamilyDomain,
			Domain:  addr,
			Port:    port,
		}, nil
	}
	// IPv4 / IPv6
	switch len(ip) {
	case net.IPv4len:
		return Address{
			Network: network,
			Family:  AddressFamilyIPv4,
			IP:      ip,
			Port:    port,
		}, nil
	case net.IPv6len:
		return Address{
			Network: network,
			Family:  AddressFamilyIPv6,
			IP:      ip,
			Port:    port,
		}, nil
	default:
		return Address{}, errors.New("invalid ip format length")
	}
}

////

func isAlphaNum(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
