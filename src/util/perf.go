package util

import (
	"encoding/binary"
	"fmt"
	"github.com/shirou/gopsutil/load"
	"math"
	"time"
)

var LoadAverage float64
var LoadAverageBytes []byte

func PrefMonitor() {
	LoadAverageBytes = make([]byte, 8)
	for {
		if stat, err := load.Avg(); err == nil {
			LoadAverage = stat.Load1
			binary.BigEndian.PutUint64(LoadAverageBytes, math.Float64bits(LoadAverage))
			fmt.Printf("[%s] load: %f\n", time.Now(), LoadAverage)
		}
		time.Sleep(5 * time.Second)
	}
}
