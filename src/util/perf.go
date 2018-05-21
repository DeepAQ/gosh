package util

import (
	"fmt"
	"github.com/shirou/gopsutil/load"
	"time"
)

var LoadAverage float64

func PrefMonitor() {
	for {
		if stat, err := load.Avg(); err == nil {
			LoadAverage = stat.Load1
			fmt.Printf("[%s] load: %f\n", time.Now(), LoadAverage)
		}
		time.Sleep(5 * time.Second)
	}
}
