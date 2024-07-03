package net

import (
	"bytes"
	"fmt"
	"net"
	"strings"
)

//LocalHostIP = IPAddress([]byte{127, 0, 0, 1})
//AnyIP = IPAddress([]byte{0, 0, 0, 0})
//LocalHostDomain = DomainAddress("localhost")
//LocalHostIPv6 = IPAddress([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
//AnyIPv6 = IPAddress([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})

type AddressFamily byte

const (
	AddressFamilyIPv4   = AddressFamily(0)
	AddressFamilyIPv6   = AddressFamily(1)
	AddressFamilyDomain = AddressFamily(2)
)

func (a AddressFamily) IsIPv4() bool {
	return a == AddressFamilyIPv4
}

func (a AddressFamily) IsIPv6() bool {
	return a == AddressFamilyIPv6
}

func (a AddressFamily) IsIP() bool {
	return a == AddressFamilyIPv4 || a == AddressFamilyIPv6
}

func (a AddressFamily) IsDomain() bool {
	return a == AddressFamilyDomain
}

type Address interface {
	IP() net.IP     // IP of this Address
	Domain() string // Domain of this Address
	Family() AddressFamily
	String() string
	Equal(Address) bool
}

func isAlphaNum(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func ParseAddress(addr string) Address {
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
	if ip != nil {
		return IPAddress(ip)
	}
	return DomainAddress(addr)
}

var bytes0 = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

func IPAddress(ip []byte) Address {
	switch len(ip) {
	case net.IPv4len:
		var addr ipv4Address = [4]byte{ip[0], ip[1], ip[2], ip[3]}
		return addr
	case net.IPv6len:
		if bytes.Equal(ip[:10], bytes0) && ip[10] == 0xff && ip[11] == 0xff {
			return IPAddress(ip[12:16])
		}
		var addr ipv6Address = [16]byte{
			ip[0], ip[1], ip[2], ip[3],
			ip[4], ip[5], ip[6], ip[7],
			ip[8], ip[9], ip[10], ip[11],
			ip[12], ip[13], ip[14], ip[15],
		}
		return addr
	default:
		panic(fmt.Sprintf("invalid ip format: %s", ip))
		return nil
	}
}

func DomainAddress(domain string) Address {
	return domainAddress(domain)
}

type ipv4Address [4]byte

func (a ipv4Address) IP() net.IP {
	return net.IP(a[:])
}

func (ipv4Address) Domain() string {
	panic("Calling Domain() on an IPv4Address.")
}

func (ipv4Address) Family() AddressFamily {
	return AddressFamilyIPv4
}

func (a ipv4Address) Equal(o Address) bool {
	if o.Family() != AddressFamilyIPv4 {
		return false
	}
	return bytes.Equal(a.IP(), o.IP())
}

func (a ipv4Address) String() string {
	return a.IP().String()
}

type ipv6Address [16]byte

func (a ipv6Address) IP() net.IP {
	return net.IP(a[:])
}

func (ipv6Address) Domain() string {
	panic("Calling Domain() on an IPv6Address.")
}

func (ipv6Address) Family() AddressFamily {
	return AddressFamilyIPv6
}

func (a ipv6Address) Equal(o Address) bool {
	if o.Family() != AddressFamilyIPv6 {
		return false
	}
	return bytes.Equal(a.IP(), o.IP())
}

func (a ipv6Address) String() string {
	return "[" + a.IP().String() + "]"
}

type domainAddress string

func (domainAddress) IP() net.IP {
	panic("Calling IP() on a DomainAddress.")
}

func (a domainAddress) Domain() string {
	return string(a)
}

func (domainAddress) Family() AddressFamily {
	return AddressFamilyDomain
}

func (a domainAddress) Equal(o Address) bool {
	if o.Family() != AddressFamilyDomain {
		return false
	}
	return strings.EqualFold(a.Domain(), o.Domain())
}

func (a domainAddress) String() string {
	return a.Domain()
}
