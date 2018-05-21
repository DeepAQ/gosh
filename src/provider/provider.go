package provider

import (
	"dubbo"
	"fmt"
	"github.com/fatih/pool"
	"github.com/valyala/fasthttp"
	"net"
	"os"
	"strconv"
	"etcd"
	"util"
)

var cp pool.Pool

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
	var err error
	cp, err = pool.NewChannelPool(0, 256, func() (net.Conn, error) {
		return net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", dubboPort))
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create channel pool:", err)
		return
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
	conn, err := cp.Get()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get connection:", err)
		ctx.Response.SetStatusCode(500)
	} else {
		defer conn.Close()
		result, err := dubbo.Invoke(&inv, conn)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invocation error:", err)
			ctx.Response.SetStatusCode(500)
		} else {
			ctx.Response.AppendBody(result)
			ctx.Response.AppendBody([]byte{0})
			ctx.Response.AppendBody(util.CpuUsageBytes)
		}
	}
}
