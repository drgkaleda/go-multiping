package pinger

import "errors"

const (
	timeSliceLength  = 8
	trackerLength    = 8
	ProtocolICMP     = 1
	ProtocolIPv6ICMP = 58
)

type ProtocolVersion int

const (
	ProtocolIpv4 = ProtocolVersion(4)
	ProtocolIpv6 = ProtocolVersion(6)
)

var (
	errInvalidConn = errors.New("invalid connection")
	errInvalidAddr = errors.New("invalid address")
)
