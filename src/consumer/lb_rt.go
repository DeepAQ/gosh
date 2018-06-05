package consumer

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"
)

func lbRT() {
	fmt.Println("Using load balancing method: Response Time")
	invokeRT = make([]int64, totalServers)
	invokeCount = make([]uint32, totalServers)

	avgRT := make([]float64, totalServers)
	actualProb := make([]float64, totalServers)
	newProb := make([]float64, totalServers)
	go func() {
		for {
			time.Sleep(5 * time.Second)
			fmt.Print("[LB_RT] count:", invokeCount)

			min := 0
			adjust := true
			for i := range avgRT {
				rt := invokeRT[i]
				count := invokeCount[i]
				actualProb[i] = float64(count)
				if count > 0 {
					avgRT[i] = math.Sqrt(float64(rt)/float64(count) + 1)
				} else {
					avgRT[i] = 0
					adjust = false
				}
				if i > 0 && avgRT[i] < avgRT[min] {
					min = i
				}
				atomic.AddInt64(&invokeRT[i], -rt)
				atomic.AddUint32(&invokeCount[i], ^(count - 1))
			}

			if adjust {
				sumProb := float64(0)
				for i := range newProb {
					newProb[i] = actualProb[i] * avgRT[min] / avgRT[i]
					sumProb += newProb[i]
				}
				for i := range newProb {
					newProb[i] /= sumProb
				}
				serverProb = newProb
			}
			fmt.Println(" avgRT:", avgRT, " prob:", serverProb)
		}
	}()
}
