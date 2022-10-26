package multiping

import (
	"net/netip"

	"github.com/drgkaleda/go-multiping/pingdata"
)

func (mp *MultiPing) batchPrepareIcmp() {
	defer close(mp.txChan)

	mp.pingData.Iterate(func(addr netip.Addr, stats *pingdata.PingStats) {
		pkt, err := mp.pinger.PrepareICMP(addr, mp.sequence)
		if err == nil {
			stats.Send(mp.sequence)
			mp.txChan <- pkt
		}
	})

}

func (mp *MultiPing) batchSendIcmp() {
	var err error
	defer mp.wg.Done()

	for pkt := range mp.txChan {
		err = mp.pinger.SendPacket(pkt)
		if err != nil {
			break
		}
	}
}
