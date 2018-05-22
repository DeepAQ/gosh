package consumer

import (
	"fmt"
	"sync/atomic"
	"time"
)

func lbAdaptive() {
	fmt.Println("Using load balancing method: Adaptive")
	serverRT = make([]int64, totalServers)
	serverRTCount = make([]uint32, totalServers)

	lastCount := make([]uint32, totalServers)
	predict := make([]uint32, totalServers)
	newProb := make([]float64, totalServers)
	go func() {
		for {
			time.Sleep(5 * time.Second)
			fmt.Print("[LB_Adaptive] count:", serverRTCount)

			sumPredict := uint32(0)
			for i, count := range serverRTCount {
				predict[i] = count + count - lastCount[i]
				if predict[i] < count {
					predict[i] = count
				}
				lastCount[i] = count
				sumPredict += predict[i]
				atomic.AddUint32(&serverRTCount[i], ^(count - 1))
			}
			fmt.Print(" predict:", predict)

			if sumPredict > 0 {
				for i := range newProb {
					newProb[i] = float64(predict[i]) / float64(sumPredict)
				}
				serverProb = newProb
			}
			fmt.Println(" prob:", serverProb)
		}
	}()
}
