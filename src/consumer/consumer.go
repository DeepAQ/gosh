package consumer

import (
	"etcd"
	"fmt"
	"github.com/valyala/fasthttp"
	"math/rand"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var client *fasthttp.Client
var servers [][]byte
var serverProb []float64
var totalServers uint64

func Start(opts map[string]string) {
	port, _ := strconv.Atoi(opts["port"])
	if port <= 0 {
		port = 20000
	}
	fmt.Println("Starting consumer agent ...")

	client = &fasthttp.Client{}

	var err error
	servers, err = etcd.Query(opts["etcd"])
	if err != nil {
		return
	}
	totalServers = uint64(len(servers))
	serverProb = make([]float64, totalServers)
	for i := range serverProb {
		serverProb[i] = 1.0 / float64(totalServers)
	}
	rand.Seed(time.Now().UnixNano())
	// Load balancing method start
	lbRT2()
	// Load balancing method end

	// Listen
	fmt.Printf("Listening on port %d\n", port)
	if err := fasthttp.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to listen:", err)
		return
	}
}

func handler(ctx *fasthttp.RequestCtx) {
	req := fasthttp.AcquireRequest()
	ctx.Request.Header.CopyTo(&req.Header)
	req.Header.SetMethod("POST")
	// Pick a server
	rand := rand.Float64()
	sum := float64(0)
	var selected int
	prob := serverProb
	for selected = 0; rand >= sum+prob[selected]; selected++ {
		sum += serverProb[selected]
	}
	req.SetHostBytes(servers[selected])

	req.SetBody(ctx.Request.Body())
	resp := fasthttp.AcquireResponse()

	beginTime := time.Now().UnixNano()
	err := client.Do(req, resp)
	rtTime := time.Now().UnixNano() - beginTime
	atomic.AddInt64(&serverRT[selected], rtTime/1E6)
	atomic.AddUint32(&serverRTCount[selected], 1)
	//fmt.Println(resp)
	if err != nil {
		ctx.Response.SetStatusCode(500)
	} else {
		ctx.Response.SetStatusCode(resp.StatusCode())
		ctx.Response.SetBody(resp.Body())
	}
}
