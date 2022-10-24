package main

import (
	"bufio"
	"flag"
	"log"
	"net/netip"
	"os"

	"github.com/drgkaleda/go-multiping"
)

func doPing(data *multiping.PingData, count int) error {
	mp, err := multiping.New(false)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		mp.Ping(data)

		// Print results
		data.Iterate(func(ip netip.Addr, val multiping.PingStats) {
			log.Println(ip, val.Valid(), val.Latency(), val.Loss(), val.Duplicate())
		})
	}
	return nil
}

func main() {
	fileName := flag.String("f", "", "File with IP list")
	count := flag.Int("c", 5, "Stop after sending count pings")

	flag.Parse()
	data := multiping.NewPingData()

	if fileName != nil && len(*fileName) > 0 {
		file, err := os.Open(*fileName)
		if err != nil {
			log.Println("Error reading file", err)
			return
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			ip, err := netip.ParseAddr(line)
			if err != nil {
				log.Println("Invalid IP", line, err)
				continue
			}

			data.Add(ip)
		}

	} else {

		if len(os.Args[1:]) == 0 {
			// No IPs configured - ping self
			data.Add(netip.MustParseAddr("127.0.0.1"))
		} else {

			for _, ipstr := range os.Args[1:] {
				ip, err := netip.ParseAddr(ipstr)
				if err != nil {
					log.Println("Invalid IP", ipstr, err)
					continue
				}

				data.Add(ip)
			}
		}
	}

	doPing(data, *count)
}
