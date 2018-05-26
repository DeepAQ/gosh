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
	lbRT()
	// Load balancing method end

	// Listen
	fmt.Printf("Listening on port %d\n", port)
	if err := fasthttp.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to listen:", err)
		return
	}
}

func handler(ctx *fasthttp.RequestCtx) {
	handlerBegin := time.Now().UnixNano()
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()

	// Pick a server
	rand := rand.Float64()
	sum := float64(0)
	var selected int
	prob := serverProb
	for selected = 0; rand >= sum+prob[selected]; selected++ {
		sum += serverProb[selected]
	}
	req.SetHostBytes(servers[selected])

	// Prepare request
	req.Header.SetMethod("POST")
	req.SetBody(ctx.Request.Body())

	serverBegin := time.Now().UnixNano()
	err := client.Do(req, resp)
	serverRT := time.Now().UnixNano() - serverBegin
	if invokeRT != nil {
		atomic.AddInt64(&invokeRT[selected], serverRT/1E3)
	}
	if invokeCount != nil {
		atomic.AddUint32(&invokeCount[selected], 1)
	}
	//fmt.Println(resp)
	ctx.Response.Header.Add("Connection", "keep-alive")
	if err != nil {
		ctx.Response.SetStatusCode(500)
	} else {
		ctx.Response.SetStatusCode(resp.StatusCode())
		ctx.Response.SetBody(resp.Body())
	}
	handlerRT := time.Now().UnixNano() - handlerBegin
	atomic.AddInt64(&consumerRT, handlerRT/1E3)
}
