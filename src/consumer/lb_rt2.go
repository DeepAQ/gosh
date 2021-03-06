package consumer

import (
	"fmt"
	"sync/atomic"
	"time"
)

func lbRT2() {
	fmt.Println("Using load balancing method: Response Time 2")
	invokeRT = make([]int64, totalServers)
	invokeCount = make([]uint32, totalServers)

	avgRT := make([]float64, totalServers)
	newProb := make([]float64, totalServers)
	go func() {
		for {
			time.Sleep(5 * time.Second)
			fmt.Print("[LB_RT2] count:", invokeCount)

			minRT := -1
			maxRT := -1
			all := true
			for i := range avgRT {
				rt := invokeRT[i]
				count := invokeCount[i]
				if count > 0 {
					avgRT[i] = float64(rt) / float64(count)
					if avgRT[i] < 1 {
						avgRT[i] = 1
					}
					if minRT < 0 || avgRT[i] < avgRT[minRT] {
						minRT = i
					}
					if maxRT < 0 || avgRT[i] > avgRT[maxRT] {
						maxRT = i
					}
				} else {
					avgRT[i] = 0
					all = false
				}
				atomic.AddInt64(&invokeRT[i], -rt)
				atomic.AddUint32(&invokeCount[i], ^(count - 1))
			}
			fmt.Print(" avgRT:", avgRT)

			sumProb := float64(0)
			minProb := 0
			for i := range newProb {
				if avgRT[i] > 0 {
					newProb[i] = serverProb[i] * avgRT[minRT] / avgRT[i]
				} else {
					newProb[i] = serverProb[i]
				}
				if i > 0 && newProb[i] < newProb[minProb] {
					minProb = i
				}
				sumProb += newProb[i]
			}
			if all && avgRT[minRT] > avgRT[maxRT]*9/10 {
				sumProb += newProb[minProb]
				newProb[minProb] *= 2
				fmt.Print(" [up] ")
			}
			for i := range newProb {
				newProb[i] /= sumProb
			}
			serverProb = newProb
			fmt.Println(" prob:", newProb)
		}
	}()
}
