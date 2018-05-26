package consumer

import (
	"fmt"
	"sync/atomic"
	"time"
)

func lbRT() {
	fmt.Println("Using load balancing method: Response Time")
	invokeRT = make([]int64, totalServers)
	invokeCount = make([]uint32, totalServers)

	avgRT := make([]float64, totalServers)
	newProb := make([]float64, totalServers)
	go func() {
		for {
			time.Sleep(5 * time.Second)
			fmt.Print("[LB_RT] count:", invokeCount)

			min := 0
			totalInvokes := uint32(0)
			for i := range avgRT {
				rt := invokeRT[i]
				count := invokeCount[i]
				totalInvokes += count
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
				atomic.AddInt64(&invokeRT[i], -rt)
				atomic.AddUint32(&invokeCount[i], ^(count - 1))
			}
			fmt.Print(" avgRT:", avgRT, " consumerRT:", float64(consumerRT)/float64(totalInvokes))
			consumerRT = 0

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
