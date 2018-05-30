package consumer

import (
	"encoding/binary"
	"fmt"
)

func lbMem() {
	fmt.Println("Using load balancing method: Memory Size")
	sumProb := float64(0)
	for i, server := range servers {
		status, body, err := server.Get(nil, "/mem")
		if err != nil || status != 200 {
			serverProb[i] = 0
		} else {
			serverProb[i] = float64(binary.BigEndian.Uint64(body))
		}
		sumProb += serverProb[i]
	}
	for i := range serverProb {
		serverProb[i] /= sumProb
	}
	fmt.Println("[LB_MEM] prob:", serverProb)
}
