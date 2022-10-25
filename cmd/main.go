package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/netip"
	"os"
	"time"

	"github.com/drgkaleda/go-multiping"
	"github.com/drgkaleda/go-multiping/pingdata"
)

const lineSep = "- - - - - - - - - - - - - - - - - - - - - - - - - -"

const (
	logLevelNone = iota
	logLevelMinimal
	logLevelFull
)

var verbose = logLevelNone
var count = 5

func doPing(data *pingdata.PingData) error {
	mp, err := multiping.New(true)
	if err != nil {
		return err
	}

	fmt.Println("Ping results:")
	if verbose == logLevelFull {
		fmt.Println(lineSep)
	}

	for i := 0; i < count; i++ {
		mp.Ping(data)

		var latencySum float32
		var lossCount uint
		var dupCount uint

		if verbose == logLevelMinimal {
			fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - -")
			fmt.Print("Ping lost: ")
		}
		// Print results
		data.Iterate(func(ip netip.Addr, val *pingdata.PingStats) {
			latencySum += val.Latency()
			if val.Loss() > 0 {
				lossCount++
			}
			if val.Duplicate() > 0 {
				dupCount++
			}

			switch verbose {
			case logLevelFull:
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
					ip, val.Latency(), val.Loss()*100, additionalInfo)
			case logLevelMinimal:
				if val.Loss() > 0 {
					fmt.Printf(" %s", ip.String())
				}
			}
		})

		switch verbose {
		case logLevelFull:
		case logLevelMinimal:
			fmt.Println()
		}

		fmt.Printf("Pinged: %d, lost: %d, avg latency: %fms, dups: %d\n",
			data.Count(), lossCount, latencySum/float32(data.Count()), dupCount)

		data.Reset()

		// Sleep before next iteration
		time.Sleep(time.Second)
	}

	return nil
}

func main() {
	fileName := flag.String("f", "", "File with IP list")
	flag.IntVar(&count, "c", 5, "Stop after sending count pings")
	flag.IntVar(&verbose, "v", logLevelNone, "Verbose logging level [0|1|2]")

	flag.Parse()
	data := pingdata.NewPingData()

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

	err := doPing(data)
	if err != nil {
		fmt.Println("Ping error", err)
	}
}
