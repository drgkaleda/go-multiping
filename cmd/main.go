package main

import (
	"log"
	"net/netip"

	"github.com/drgkaleda/go-multiping"
)

func main() {
	data := multiping.NewPingData()
	mp, err := multiping.New(false)
	if err != nil {
		log.Println(err)
		return
	}

	data.Add(netip.MustParseAddr("1.1.1.1"))
	data.Add(netip.MustParseAddr("8.8.8.8"))
	data.Add(netip.MustParseAddr("4.4.4.4"))
	// For unknown reasons this IP reports dups. Usefull for testing
	data.Add(netip.MustParseAddr("74.3.163.56"))

	for i := 0; i < 100; i++ {
		mp.Ping(data)

		data.Iterate(func(ip netip.Addr, val multiping.PingStats) {
			log.Println(ip, val.Valid(), val.Latency(), val.Loss(), val.Duplicate())
		})
	}
}
