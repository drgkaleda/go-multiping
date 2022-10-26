// This file code is based on https://github.com/go-ping/ping
package pinger

import (
	"math/rand"
	"net/netip"

	"golang.org/x/net/icmp"
)

// NewPinger returns a new Pinger instance
func NewPinger(network, protocol string, id uint16) *Pinger {
	p := &Pinger{
		Size: timeSliceLength,

		id:       id,
		ipaddr:   nil,
		network:  network,
		protocol: protocol,

		Tracker: int64(rand.Uint64()),
	}
	return p
}

// Pinger represents a packet sender.
type Pinger struct {
	// Size of packet being sent
	Size int

	// Tracker: Used to uniquely identify packet when non-priviledged
	Tracker int64

	ipaddr *netip.Addr

	id uint16
	// network is one of "ip", "ip4", or "ip6".
	network string
	// protocol is "icmp" or "udp".
	protocol string

	//conn4 is ipv4 icmp PacketConn
	conn4 *icmp.PacketConn

	//conn6 is ipv6 icmp PacketConn
	conn6 *icmp.PacketConn
}

// SetIPAddr sets the ip address of the target host.
func (p *Pinger) SetIPAddr(ipaddr *netip.Addr) {
	p.ipaddr = ipaddr
}

// IPAddr returns the ip address of the target host.
func (p *Pinger) IPAddr() *netip.Addr {
	return p.ipaddr
}

// Addr returns the string ip address of the target host.
func (p *Pinger) Addr() string {
	if p.ipaddr != nil {
		return p.ipaddr.String()
	}
	return ""
}

// SetConns setups IPv4 and IPv6 connections to pinger
func (p *Pinger) SetConns(c4 *icmp.PacketConn, c6 *icmp.PacketConn) {
	p.conn4 = c4
	p.conn6 = c6
}

// SetPrivileged sets the type of ping pinger will send.
// false means pinger will send an "unprivileged" UDP ping.
// true means pinger will send a "privileged" raw ICMP ping.
// NOTE: setting to true requires that it be run with super-user privileges.
func (p *Pinger) SetPrivileged(privileged bool) {
	if privileged {
		p.protocol = "icmp"
	} else {
		p.protocol = "udp"
	}
}

// Privileged returns whether pinger is running in privileged mode.
func (p *Pinger) Privileged() bool {
	return p.protocol == "icmp"
}
