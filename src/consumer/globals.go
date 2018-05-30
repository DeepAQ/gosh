package consumer

import "github.com/valyala/fasthttp"

var client *fasthttp.Client

var servers []*fasthttp.HostClient
var serverProb []float64
var totalServers int

var invokeCount []uint32
var invokeRT []int64
var consumerRT int64
