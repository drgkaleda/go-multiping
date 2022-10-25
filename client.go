package multiping

import "github.com/drgkaleda/go-multiping/pingdata"

// Unified interface to process ping data
type PingClient interface {
	PingProcess(pr *pingdata.PingData)
}
