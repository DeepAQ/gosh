package consumer

import "github.com/valyala/fasthttp"

var client *fasthttp.Client

var servers [][]byte
var serverProb []float64
var totalServers uint64

var invokeCount []uint32
var invokeRT []int64
var consumerRT int64
