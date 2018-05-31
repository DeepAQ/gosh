package provider

import (
	"dubbo"
	"etcd"
	"fmt"
	"github.com/fatih/pool"
	"github.com/valyala/fasthttp"
	"net"
	"strconv"
	"unsafe"
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
	cp, err = pool.NewChannelPool(0, 200, func() (net.Conn, error) {
		return net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", dubboPort))
	})
	if err != nil {
		fmt.Println("Failed to create channel pool:", err)
		return
	}

	// Create buffer pool
	util.BufPool = util.NewBufferPool(256)

	// Register to etcd
	etcd.Register(opts["etcd"], port)

	// Start performance monitor
	//go util.PrefMonitor()

	// Listen
	fmt.Printf("Listening on port %d, dubbo port %d\n", port, dubboPort)
	if err := fasthttp.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		fmt.Println("Failed to listen:", err)
		return
	}
}

func handler(ctx *fasthttp.RequestCtx) {
	//path := string(ctx.Path())
	//if path == "/perf" {
	//	var perfBytes [8]byte
	//	binary.BigEndian.PutUint64(perfBytes[:], math.Float64bits(util.LoadAverage))
	//	ctx.Response.AppendBody(perfBytes[:])
	//} else if path == "/mem" {
	//	var memBytes [8]byte
	//	binary.BigEndian.PutUint64(memBytes[:], util.TotalMem())
	//	ctx.Response.AppendBody(memBytes[:])
	//} else {
	inv := &dubbo.Invocation{
		DubboVersion: "2.0.0",
	}
	ctx.Request.PostArgs().VisitAll(func(k, v []byte) {
		switch *(*string)(unsafe.Pointer(&k)) {
		case "interface":
			inv.ServiceName = v
		case "method":
			inv.MethodName = v
		case "parameterTypesString":
			inv.MethodParamTypes = v
		case "parameter":
			inv.MethodArgs = v
		}
	})
	conn, err := cp.Get()
	if err != nil {
		fmt.Println("Failed to get connection:", err)
		ctx.Response.SetStatusCode(500)
	} else {
		result, err := dubbo.Invoke(inv, conn)
		conn.Close()
		if err != nil {
			fmt.Println("Invocation error:", err)
			ctx.Response.SetStatusCode(500)
		} else {
			ctx.Response.SetBody(result)
		}
	}
	//}
}
