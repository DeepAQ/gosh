package consumer

import (
	"etcd"
	"fmt"
	"github.com/valyala/fasthttp"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"
	"unsafe"
)

func Start(opts map[string]string) {
	port, _ := strconv.Atoi(opts["port"])
	if port <= 0 {
		port = 20000
	}
	fmt.Println("Starting consumer agent ...")

	hosts, err := etcd.Query(opts["etcd"])
	if err != nil {
		return
	}
	totalServers = len(hosts)
	servers = make([]fasthttp.HostClient, totalServers)
	for i, host := range hosts {
		servers[i] = fasthttp.HostClient{
			Addr:                          *(*string)(unsafe.Pointer(&host)),
			MaxConns:                      256,
			MaxIdleConnDuration:           60 * time.Second,
			ReadBufferSize:                1024,
			WriteBufferSize:               1024,
			DisableHeaderNamesNormalizing: true,
		}
	}

	serverProb = make([]float64, totalServers)
	for i := range serverProb {
		serverProb[i] = 1.0 / float64(totalServers)
	}
	rand.Seed(time.Now().UnixNano())
	// Load balancing method start
	lbRT()
	// Load balancing method end

	// Listen
	fmt.Printf("Listening on port %d\n", port)
	if err := fasthttp.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		fmt.Println("Failed to listen:", err)
		return
	}
}

func handler(ctx *fasthttp.RequestCtx) {
	handlerBegin := time.Now().UnixNano()

	// Pick a server
	rand := rand.Float64()
	sum := float64(0)
	var selected int
	prob := serverProb
	for selected = 0; rand >= sum+prob[selected]; selected++ {
		sum += serverProb[selected]
	}

	// Prepare request
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.Header.SetMethod("POST")
	req.Header.SetHost(servers[selected].Addr)
	req.SetBody(ctx.Request.Body())

	serverBegin := time.Now().UnixNano()
	err := servers[selected].Do(req, resp)
	serverRT := time.Now().UnixNano() - serverBegin

	fasthttp.ReleaseRequest(req)
	if err != nil {
		ctx.Response.SetStatusCode(500)
	} else {
		ctx.Response.SetStatusCode(resp.StatusCode())
		ctx.Response.SetBody(resp.Body())
	}
	fasthttp.ReleaseResponse(resp)

	if invokeRT != nil {
		atomic.AddInt64(&invokeRT[selected], serverRT/1E3)
	}
	if invokeCount != nil {
		atomic.AddUint32(&invokeCount[selected], 1)
	}
	handlerRT := time.Now().UnixNano() - handlerBegin
	atomic.AddInt64(&consumerRT, handlerRT/1E3)
}
