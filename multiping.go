package multiping

/**
 *    ***   The motivation for this multi-ping fork   ***
 *
 * There are quite a few Go pinger, but all of them have issues:
 *  * https://github.com/go-ping/ping works fine, but has problems when running
 *    several pingers in goroutines. When pinging ~300 hosts it looses ~1/3 packets.
 *  * https://github.com/caucy/batch_ping is umaintened for a long time and did not work for me at all.
 *  * https://github.com/rosenlo/go-MultiPing is a very young fork, has issues with logger, some parts
 *    of code are ineffective.
 *
 *  Also need to mention that all these pingers are periodic pingers, they try to mimmic shell ping command.
 * They run in internal loop, cancel that loop after timeout. They *can* be used, but you have to adjust your
 * code to their style. Instead I wanted a pinger, that can ping multipple hosts at a time and be robust.
 * I don't think its a problem for ping user to run it in a loop and don't want any hidden logic.
 * So this ping is loosely based on above mentioned projects. It can ping multipple clients.
 * And is cholesterol free.
 **/

import (
	"context"
	"math/rand"
	"net/netip"
	"sync"
	"time"

	"github.com/drgkaleda/go-multiping/pingdata"
	"github.com/drgkaleda/go-multiping/pinger"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

var (
	ipv4Proto = map[string]string{"icmp": "ip4:icmp", "udp": "udp4"}
	ipv6Proto = map[string]string{"icmp": "ip6:ipv6-icmp", "udp": "udp6"}
)

type MultiPing struct {
	// Locks MultiPing to protect internal members
	sync.RWMutex

	// Sync internal goroutines
	wg sync.WaitGroup

	// Timeout specifies a timeout before ping exits, regardless of how many
	// packets have been received. Default is 1s.
	Timeout time.Duration

	// Tracker: Used to uniquely identify packet when non-priviledged
	Tracker int64

	ctx    context.Context    // context for timeouting
	cancel context.CancelFunc // Do I need it ?

	pinger   *pinger.Pinger
	pingData *pingdata.PingData

	id       uint16
	sequence uint16 // ICMP seq number. Incremented on every ping
	network  string // one of "ip", "ip4", or "ip6"
	protocol string // protocol is "icmp" or "udp".
	conn4    *icmp.PacketConn
	conn6    *icmp.PacketConn
	rxChan   chan *pinger.Packet
}

func New(privileged bool) (*MultiPing, error) {
	protocol := "udp"
	if privileged {
		protocol = "icmp"
	}

	rand.Seed(time.Now().UnixNano())
	mp := &MultiPing{
		Timeout:  time.Second,
		id:       uint16(rand.Intn(0xffff)),
		network:  "ip",
		protocol: protocol,
		Tracker:  rand.Int63(),
	}

	mp.pinger = pinger.NewPinger(mp.network, mp.protocol, mp.id)
	mp.pinger.SetPrivileged(privileged)
	mp.pinger.Tracker = mp.Tracker

	// try initialise connections to test that everything's working
	err := mp.restart()
	if err != nil {
		mp.close()
		return nil, err
	}

	// Sequence counter. It will be incremented in mp.restart on every ping
	// Start with quite big initial value, so overwrap will occure fast (easier debugin)
	mp.sequence = 0xfff0

	return mp, nil
}

func (mp *MultiPing) restart() (err error) {
	// ipv4
	mp.conn4, err = icmp.ListenPacket(ipv4Proto[mp.protocol], "")
	if err != nil {
		return err
	}
	err = mp.conn4.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL, true)
	if err != nil {
		return err
	}

	// ipv6 (note IPv6 may be disabled on OS and may fail)
	mp.conn6, err = icmp.ListenPacket(ipv6Proto[mp.protocol], "")
	if err == nil {
		mp.conn6.IPv6PacketConn().SetControlMessage(ipv6.FlagHopLimit, true)
	}

	mp.pinger.SetConns(mp.conn4, mp.conn6)
	mp.sequence++
	// I use zero sequence number in statistics struct
	// to detect duplicates, thus don't use it as valid sequence number
	if mp.sequence == 0 {
		mp.sequence++
	}

	mp.rxChan = make(chan *pinger.Packet)

	return nil
}

// closes active connections
func (mp *MultiPing) close() {
	if mp.conn4 != nil {
		mp.conn4.Close()
	}
	if mp.conn6 != nil {
		mp.conn6.Close()
	}
}

// cleanup cannot be done in close, because some goroutines may be using struct members
func (mp *MultiPing) cleanup() {
	// Close channels
	close(mp.rxChan)

	// invalidate connections
	mp.conn4 = nil
	mp.conn6 = nil
	mp.pinger.SetConns(nil, nil)

	// Invalidate pingData pointer (prevent from possible data corruption in future)
	mp.pingData = nil
	// Invalidate IP address
	mp.pinger.SetIPAddr(nil)
}

// Ping is blocking function and runs for mp.Timeout time and pings all hosts in data
func (mp *MultiPing) Ping(data *pingdata.PingData) {
	if data.Count() == 0 {
		return
	}

	// Lock the pinger - its instance may be reused by several clients
	mp.Lock()
	defer mp.Unlock()

	err := mp.restart()
	if err != nil {
		return
	}

	// Some subfunctions in goroutines will need this pointer to store ping results
	mp.pingData = data

	mp.ctx, mp.cancel = context.WithTimeout(context.Background(), mp.Timeout)
	defer mp.cancel()

	// This goroutine depends on rxChan and no need to add it to workgroup
	// It will terminate on channel close
	go mp.batchProcessPacket()

	if mp.conn4 != nil {
		mp.wg.Add(1)
		mp.conn4.SetReadDeadline(time.Now().Add(mp.Timeout))
		go mp.batchRecvICMP(pinger.ProtocolIpv4)
	}
	if mp.conn6 != nil {
		mp.wg.Add(1)
		mp.conn6.SetReadDeadline(time.Now().Add(mp.Timeout))
		go mp.batchRecvICMP(pinger.ProtocolIpv6)
	}

	// Sender goroutine
	mp.wg.Add(1)
	go func() {
		defer mp.wg.Done()
		mp.pingData.Iterate(func(addr netip.Addr, stats *pingdata.PingStats) {
			mp.pinger.SetIPAddr(&addr)
			stats.Send(mp.sequence)

			mp.pinger.SendICMP(mp.sequence)
			time.Sleep(time.Millisecond)
		})
	}()

	// wait for timeout and close connections
	<-mp.ctx.Done()
	mp.close()

	// wait for all goroutines to terminate
	mp.wg.Wait()

	mp.cleanup()
}

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

		if stats, ok := mp.pingData.Get(recv.Src); ok {
			stats.Recv(pingStats.Seq, pingStats.RTT)
		}
	}
}
