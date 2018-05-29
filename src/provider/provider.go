package provider

import (
	"bytes"
	"dubbo"
	"etcd"
	"fmt"
	"github.com/fatih/pool"
	"github.com/valyala/fasthttp"
	"net"
	"os"
	"strconv"
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
		fmt.Fprintln(os.Stderr, "Failed to create channel pool:", err)
		return
	}

	// Register to etcd
	etcd.Register(opts["etcd"], port)

	// Start performance monitor
	//go util.PrefMonitor()

	// Listen
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to listen:", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Listening on port %d, dubbo port %d\n", port, dubboPort)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to accept new connection:", err)
		}
		go rawHandler(conn)
	}
}

func rawHandler(conn net.Conn) {
	defer conn.Close()
	for {
		var data [1024]byte
		limit := 0
		for {
			i, err := conn.Read(data[limit:])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to read from client:", err)
				return
			}
			limit += i
			if data[limit-2] == 0xbe && data[limit-1] == 0xef {
				break
			}
		}

		go func() {
			i1 := bytes.IndexByte(data[:], 0xff)
			i2 := i1 + 1 + bytes.IndexByte(data[i1+1:], 0xff)
			i3 := i2 + 1 + bytes.IndexByte(data[i2+1:], 0xff)
			if i1 < 0 || i2 < 0 || i3 < 0 {
				fmt.Fprintln(os.Stderr, "bad request")
				return
			}

			inv := &dubbo.Invocation{
				DubboVersion:     "2.0.0",
				ServiceName:      data[2:i1],
				MethodName:       data[i1+1 : i2],
				MethodParamTypes: data[i2+1 : i3],
				MethodArgs:       data[i3+1 : limit-2],
			}
			sConn, err := cp.Get()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to get connection:", err)
			} else {
				result, err := dubbo.Invoke(inv, sConn)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Invocation error:", err)
				} else {
					conn.Write(result)
				}
				sConn.Close()
			}
		}()
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
	args := ctx.PostArgs()
	inv := &dubbo.Invocation{
		DubboVersion:     "2.0.0",
		ServiceName:      args.Peek("interface"),
		MethodName:       args.Peek("method"),
		MethodParamTypes: args.Peek("parameterTypesString"),
		MethodArgs:       args.Peek("parameter"),
	}
	conn, err := cp.Get()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get connection:", err)
		ctx.Response.SetStatusCode(500)
	} else {
		defer conn.Close()
		result, err := dubbo.Invoke(inv, conn)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invocation error:", err)
			ctx.Response.SetStatusCode(500)
		} else {
			ctx.Response.AppendBody(result)
		}
	}
	//}
}
