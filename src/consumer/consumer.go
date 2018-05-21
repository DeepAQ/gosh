package consumer

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"math/rand"
	"os"
	"strconv"
	"time"
	"etcd"
	"bytes"
	"encoding/binary"
	"math"
)

var client *fasthttp.Client
var servers [][]byte
var serverLoad []float64
var serverProb []float64
var total uint64

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
	total = uint64(len(servers))
	serverLoad = make([]float64, total)
	serverProb = make([]float64, total)
	for i := range serverProb {
		serverProb[i] = 1.0 / float64(total)
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
	req.SetHostBytes(servers[selected])

	req.SetBody(ctx.Request.Body())
	resp := fasthttp.AcquireResponse()
	err := client.Do(req, resp)
	//fmt.Println(resp)
	if err != nil {
		ctx.Response.SetStatusCode(500)
	} else {
		ctx.Response.SetStatusCode(resp.StatusCode())
		if resp.StatusCode() == 200 {
			body := resp.Body()
			index := bytes.IndexByte(body, 0)
			serverLoad[selected] = math.Float64frombits(binary.BigEndian.Uint64(body[index+1:]))
			ctx.Response.SetBody(body[:index])
		}
	}
}

func lb() {
	realLoad := make([]float64, total)
	newProb := make([]float64, total)
	for {
		time.Sleep(5 * time.Second)
		max := 0
		for i := range serverLoad {
			realLoad[i] = serverLoad[i]
			serverLoad[i] = 1
			if realLoad[i] < 0.01 {
				realLoad[i] = 0.01
			}
			if i > 0 && realLoad[i] > realLoad[max] {
				max = i
			}
		}
		sumLoad := float64(0)
		for i := range newProb {
			newProb[i] = serverProb[i] * realLoad[max] / realLoad[i]
			sumLoad += newProb[i]
		}
		for i := range newProb {
			newProb[i] /= sumLoad
		}
		serverProb = newProb
		fmt.Println("[LB] load:", realLoad, "prob:", newProb)
	}
}
