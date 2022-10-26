package multiping

import "github.com/drgkaleda/go-multiping/pinger"

func (mp *MultiPing) batchRecvICMP(proto pinger.ProtocolVersion) {
	defer func() {
		mp.wg.Done()
	}()

	for {
		pkt, err := mp.pinger.RecvPacket(proto)
		if err != nil {
			return
		}

		mp.rxChan <- pkt
	}
}

// This function runs in goroutine and nobody is interested in return errors
// Discard errors silently
func (mp *MultiPing) batchProcessPacket() {
	for recv := range mp.rxChan {
		pingStats := mp.pinger.ParsePacket(recv)
		if pingStats.Tracker != mp.Tracker {
			continue
		}

		if stats, ok := mp.pingData.Get(recv.Addr); ok {
			stats.Recv(pingStats.Seq, pingStats.RTT)
		}
	}
}
