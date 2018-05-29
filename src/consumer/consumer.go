package consumer

import (
	"bytes"
	"etcd"
	"fmt"
	"github.com/fatih/pool"
	"github.com/valyala/fasthttp"
	"math/rand"
	"net"
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

	var err error
	servers, err = etcd.Query(opts["etcd"])
	if err != nil {
		return
	}
	totalServers = uint64(len(servers))
	serverPool = make([]pool.Pool, totalServers)
	serverProb = make([]float64, totalServers)
	for i, s := range servers {
		serverProb[i] = 1.0 / float64(totalServers)
		serverPool[i], err = pool.NewChannelPool(0, 200, func() (net.Conn, error) {
			return net.Dial("tcp", string(s))
		})
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
	args := ctx.PostArgs()
	buf := bytes.Buffer{}
	buf.WriteByte(0xde)
	buf.WriteByte(0xad)
	buf.Write(args.Peek("interface"))
	buf.WriteByte(0xff)
	buf.Write(args.Peek("method"))
	buf.WriteByte(0xff)
	buf.Write(args.Peek("parameterTypesString"))
	buf.WriteByte(0xff)
	buf.Write(args.Peek("parameter"))
	buf.WriteByte(0xbe)
	buf.WriteByte(0xef)

	conn, err := serverPool[selected].Get()
	if err != nil {
		fmt.Println("Failed to get connection:", err)
		ctx.Response.SetStatusCode(500)
		return
	}
	serverBegin := time.Now().UnixNano()
	if _, err := conn.Write(buf.Bytes()); err != nil {
		fmt.Println("Failed to write:", err)
		ctx.Response.SetStatusCode(500)
		conn.Close()
		return
	}
	var result [1024]byte
	limit, err := conn.Read(result[:])
	if err != nil {
		fmt.Println("Failed to read:", err)
		ctx.Response.SetStatusCode(500)
		conn.Close()
		return
	}
	serverRT := time.Now().UnixNano() - serverBegin
	conn.Close()
	ctx.Response.SetBody(result[:limit])

	if invokeRT != nil {
		atomic.AddInt64(&invokeRT[selected], serverRT/1E3)
	}
	if invokeCount != nil {
		atomic.AddUint32(&invokeCount[selected], 1)
	}
	handlerRT := time.Now().UnixNano() - handlerBegin
	atomic.AddInt64(&consumerRT, handlerRT/1E3)
}
