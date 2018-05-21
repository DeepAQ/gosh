package util

import (
	"fmt"
	"time"
	"github.com/shirou/gopsutil/cpu"
	"encoding/binary"
	"math"
)

var lastIdle, lastTotal, CpuUsage float64
var CpuUsageBytes []byte

func PrefMonitor() {
	CpuUsageBytes = make([]byte, 8)
	for {
		if stat, err := cpu.Times(false); err == nil && len(stat) > 0 {
			curStat := stat[0]
			totalTicks := curStat.User + curStat.Nice + curStat.System + curStat.Idle + curStat.Iowait + curStat.Irq + curStat.Softirq
			if lastTotal > 0 {
				deltaTicks := totalTicks - lastTotal
				deltaIdle := curStat.Idle - lastIdle
				CpuUsage = 1.0 - deltaIdle/deltaTicks
				binary.BigEndian.PutUint64(CpuUsageBytes, math.Float64bits(CpuUsage))
				fmt.Printf("[%s] cpu: %f\n", time.Now(), CpuUsage)
			}
			lastTotal = totalTicks
			lastIdle = curStat.Idle
		}
		time.Sleep(5 * time.Second)
	}
}
