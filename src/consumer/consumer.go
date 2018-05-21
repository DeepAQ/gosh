package consumer

import (
	"fmt"
	"strconv"
	"github.com/valyala/fasthttp"
	"os"
	"github.com/coreos/etcd/clientv3"
	"time"
	"context"
	"math/rand"
	"sync/atomic"
)

var client *fasthttp.Client
var servers [][]byte
var current, total uint64

func Start(opts map[string]string) {
	port, _ := strconv.Atoi(opts["port"])
	if port <= 0 {
		port = 20000
	}
	fmt.Println("Starting consumer agent ...")

	client = &fasthttp.Client{}

	etcd := opts["etcd"]
	fmt.Printf("Querying from etcd %s\n", etcd)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcd},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create etcd client:", err)
		return
	}
	resp, err := cli.Get(context.TODO(), "dubbomesh/", clientv3.WithPrefix())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to query etcd:", err)
		return
	}
	fmt.Println("Providers:", resp.Kvs)
	total = uint64(len(resp.Kvs))
	servers = make([][]byte, total)
	for i, kv := range resp.Kvs {
		servers[i] = kv.Value
	}
	rand.Seed(time.Now().UnixNano())

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
	req.SetHostBytes(servers[atomic.AddUint64(&current, 1) % total])
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
