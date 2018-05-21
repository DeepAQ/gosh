package consumer

import (
	"encoding/binary"
	"etcd"
	"fmt"
	"github.com/valyala/fasthttp"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

var client *fasthttp.Client
var servers [][]byte
var serverProb []float64
var serverCount []uint32
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
	serverCount = make([]uint32, totalServers)
	for i := range serverProb {
		serverProb[i] = 1.0 / float64(totalServers)
	}
	rand.Seed(time.Now().UnixNano())
	go lb()

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
	for selected = 0; rand >= sum+serverProb[selected]; selected++ {
		sum += serverProb[selected]
	}
	atomic.AddUint32(&serverCount[selected], 1)
	req.SetHostBytes(servers[selected])

	req.SetBody(ctx.Request.Body())
	resp := fasthttp.AcquireResponse()
	err := client.Do(req, resp)
	//fmt.Println(resp)
	if err != nil {
		ctx.Response.SetStatusCode(500)
	} else {
		ctx.Response.SetStatusCode(resp.StatusCode())
		ctx.Response.SetBody(resp.Body())
	}
}

func lb() {
	serverLoad := make([]float64, totalServers)
	newProb := make([]float64, totalServers)
	for {
		time.Sleep(5 * time.Second)
		max := 0
		for i, server := range servers {
			status, body, err := client.Get(nil, "http://"+string(server)+"/perf")
			if err != nil || status != 200 {
				serverLoad[i] = 1
			} else {
				serverLoad[i] = math.Float64frombits(binary.BigEndian.Uint64(body))
			}
			if serverLoad[i] < 0.01 {
				serverLoad[i] = 0.01
			}
			if i > 0 && serverLoad[i] > serverLoad[max] {
				max = i
			}
		}
		sumLoad := float64(0)
		for i := range newProb {
			newProb[i] = serverProb[i] * serverLoad[max] / serverLoad[i]
			sumLoad += newProb[i]
		}
		for i := range newProb {
			newProb[i] /= sumLoad
		}
		serverProb = newProb
		fmt.Println("[LB] load:", serverLoad, "prob:", newProb, "count:", serverCount)
	}
}
