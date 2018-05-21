package util

import (
	"fmt"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
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

func TotalMem() uint64 {
	if stat, err := mem.VirtualMemory(); err == nil {
		return stat.Total
	}
	return 0
}
