package pinger

import (
	"net/netip"
	"time"
)

type Packet struct {
	Bytes []byte
	Len   int
	TTL   int
	Proto ProtocolVersion
	Src   netip.Addr
}

type IcmpStats struct {
	Valid   bool
	RTT     time.Duration
	Tracker int64
	Seq     uint16
}
