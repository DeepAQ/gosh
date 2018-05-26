package consumer

import (
	"fmt"
	"sync/atomic"
	"time"
)

func lbAdaptive() {
	fmt.Println("Using load balancing method: Adaptive")
	invokeRT = make([]int64, totalServers)
	invokeCount = make([]uint32, totalServers)

	lastCount := make([]uint32, totalServers)
	predict := make([]int, totalServers)
	newProb := make([]float64, totalServers)
	go func() {
		for {
			time.Sleep(5 * time.Second)
			fmt.Print("[LB_Adaptive] count:", invokeCount)

			sumPredict := 0
			for i, count := range invokeCount {
				predict[i] = int(count+count) - int(lastCount[i]) + 1
				if predict[i] < int(count) {
					predict[i] = int(count)
				}
				lastCount[i] = count
				sumPredict += predict[i]
				atomic.AddUint32(&invokeCount[i], ^(count - 1))
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
