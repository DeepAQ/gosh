package consumer

import "github.com/fatih/pool"

var servers [][]byte
var serverPool []pool.Pool
var serverProb []float64
var totalServers uint64

var invokeCount []uint32
var invokeRT []int64
var consumerRT int64
