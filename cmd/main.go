package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/netip"
	"os"

	"github.com/drgkaleda/go-multiping"
)

const lineSep = "- - - - - - - - - - - - - - - - - - - - - - - - - -"

func doPing(data *multiping.PingData, count int) error {
	mp, err := multiping.New(false)
	if err != nil {
		return err
	}

	fmt.Println("Ping results:")
	fmt.Println(lineSep)
	for i := 0; i < count; i++ {
		mp.Ping(data)

		// Print results
		data.Iterate(func(ip netip.Addr, val multiping.PingStats) {
			var additionalInfo string
			if !val.Valid() || val.Duplicate() > 0 {
				additionalInfo = "("
				if !val.Valid() {
					additionalInfo = additionalInfo + " invalid "
				}
				if val.Duplicate() > 0 {
					additionalInfo = additionalInfo + fmt.Sprintf(" dupps=%d ", val.Duplicate())
				}
			}
			fmt.Printf("%16s\t%fms\t%f%%\t%s\n",
				ip, val.Latency(), val.Loss(), additionalInfo)
		})
		fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - -")
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
