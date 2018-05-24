package provider

import (
	"dubbo"
	"encoding/binary"
	"etcd"
	"fmt"
	"github.com/valyala/fasthttp"
	"math"
	"os"
	"strconv"
	"time"
	"util"
)

//var cp pool.Pool

func Start(opts map[string]string) {
	port, _ := strconv.Atoi(opts["port"])
	if port <= 0 {
		port = 30000
	}
	dubboPort, _ := strconv.Atoi(opts["dubbo.port"])
	if dubboPort <= 0 {
		dubboPort = 20880
	}
	fmt.Println("Starting provider agent ...")

	// Create channel pool
	//var err error
	//cp, err = pool.NewChannelPool(0, 200, func() (net.Conn, error) {
	//	return net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", dubboPort))
	//})
	//if err != nil {
	//	fmt.Fprintln(os.Stderr, "Failed to create channel pool:", err)
	//	return
	//}
	for {
		if err := dubbo.Connect(fmt.Sprintf("127.0.0.1:%d", dubboPort), 8); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// Register to etcd
	etcd.Register(opts["etcd"], port)

	// Start performance monitor
	go util.PrefMonitor()

	// Listen
	fmt.Printf("Listening on port %d, dubbo port %d\n", port, dubboPort)
	if err := fasthttp.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to listen:", err)
		return
	}
}

func handler(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	if path == "/perf" {
		perfBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(perfBytes, math.Float64bits(util.LoadAverage))
		ctx.Response.AppendBody(perfBytes)
	} else if path == "/mem" {
		memBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(memBytes, util.TotalMem())
		ctx.Response.AppendBody(memBytes)
	} else {
		args := ctx.PostArgs()
		inv := dubbo.Invocation{
			DubboVersion:     "2.0.0",
			ServiceName:      string(args.Peek("interface")),
			ServiceVersion:   "",
			MethodName:       string(args.Peek("method")),
			MethodParamTypes: string(args.Peek("parameterTypesString")),
			MethodArgs:       string(args.Peek("parameter")),
			Attachments:      make(map[string]string),
		}
		inv.Attachments["path"] = inv.ServiceName
		result := dubbo.Invoke(&inv)
		if result == nil {
			fmt.Fprintln(os.Stderr, "Invocation error")
			ctx.Response.SetStatusCode(500)
		} else {
			ctx.Response.AppendBody(result)
		}
	}
}
