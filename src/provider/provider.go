package provider

import (
	"dubbo"
	"etcd"
	"fmt"
	"github.com/valyala/fasthttp"
	"strconv"
	"unsafe"
	"util"
)

func Start(opts map[string]string) {
	port, _ := strconv.Atoi(opts["port"])
	if port <= 0 {
		port = 30000
	}
	dubboPort, _ := strconv.Atoi(opts["dubbo.port"])
	if dubboPort <= 0 {
		dubboPort = 20880
	}
	weight, _ := strconv.Atoi(opts["weight"])
	if weight <= 0 {
		weight = 1
	}
	fmt.Println("Starting provider agent ...")

	// Init pools and semaphore
	util.InitPools(16, 64, []string{fmt.Sprintf("127.0.0.1:%d", dubboPort)})
	util.InitSem(199)

	// Register to etcd
	etcd.Register(opts["etcd"], port, weight)

	// Start performance monitor
	//go util.PrefMonitor()

	fmt.Printf("Listening on port %d, dubbo port %d\n", port, dubboPort)
	// Listen HTTP
	//if err := fasthttp.ListenAndServe(fmt.Sprintf(":%d", port), httpHandler); err != nil {
	//	fmt.Println("Failed to listen:", err)
	//	return
	//}

	// Listen Cafe
	listenCafe(port)
}

func httpHandler(ctx *fasthttp.RequestCtx) {
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
	inv := dubbo.Invocation{
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
	result, err := dubbo.Invoke(inv)
	if err != nil || result == nil {
		fmt.Println("Invocation error:", err)
		ctx.Response.SetStatusCode(500)
	} else {
		ctx.Response.SetBody(result)
	}
	//}
}
