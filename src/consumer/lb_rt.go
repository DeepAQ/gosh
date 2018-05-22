package consumer

import (
	"fmt"
	"sync/atomic"
	"time"
)

var serverRT []int64
var serverRTCount []uint32

func lbRT() {
	fmt.Println("Using load balancing method: Response Time")
	serverRT = make([]int64, totalServers)
	serverRTCount = make([]uint32, totalServers)

	avgRT := make([]float64, totalServers)
	newProb := make([]float64, totalServers)
	go func() {
		for {
			time.Sleep(2 * time.Second)
			fmt.Print("[LB_RT] count:", serverRTCount)

			min := 0
			for i := range avgRT {
				rt := serverRT[i]
				count := serverRTCount[i]
				if count > 0 {
					avgRT[i] = float64(rt) / float64(count)
					if avgRT[i] < 1 {
						avgRT[i] = 1
					}
				} else {
					avgRT[i] = 1
				}
				if i > 0 && avgRT[i] < avgRT[min] {
					min = i
				}
				atomic.AddInt64(&serverRT[i], -rt)
				atomic.AddUint32(&serverRTCount[i], ^(count - 1))
			}
			fmt.Print(" avgRT:", avgRT)

			sumProb := float64(0)
			for i := range newProb {
				newProb[i] = serverProb[i] * avgRT[min] / avgRT[i]
				sumProb += newProb[i]
			}
			for i := range newProb {
				newProb[i] /= sumProb
			}
			serverProb = newProb
			fmt.Println(" prob:", newProb)
		}
	}()
}
