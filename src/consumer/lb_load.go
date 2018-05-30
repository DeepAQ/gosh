package consumer

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

func lbLoad() {
	fmt.Println("Using load balancing method: Load Average")
	serverLoad := make([]float64, totalServers)
	newProb := make([]float64, totalServers)
	go func() {
		for {
			time.Sleep(5 * time.Second)
			max := 0
			for i, server := range servers {
				status, body, err := server.Get(nil, "/perf")
				if err != nil || status != 200 {
					serverLoad[i] = 1
				} else {
					serverLoad[i] = math.Float64frombits(binary.BigEndian.Uint64(body))
				}
				if serverLoad[i] < 0.01 {
					serverLoad[i] = 0.01
				}
				if i > 0 && serverLoad[i] > serverLoad[max] {
					max = i
				}
			}
			sumLoad := float64(0)
			for i := range newProb {
				newProb[i] = serverProb[i] * serverLoad[max] / serverLoad[i]
				sumLoad += newProb[i]
			}
			for i := range newProb {
				newProb[i] /= sumLoad
			}
			serverProb = newProb
			fmt.Println("[LB] load:", serverLoad, "prob:", newProb)
		}
	}()
}
