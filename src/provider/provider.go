package provider

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"dubbo"
	"github.com/fatih/pool"
	"net"
	"os"
	"github.com/coreos/etcd/clientv3"
	"time"
	"context"
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
	cp, err = pool.NewChannelPool(0, 100, func() (net.Conn, error) {
		return net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", dubboPort))
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create channel pool:", err)
		return
	}

	// Register to etcd
	register(opts["etcd"], port)

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
			ctx.Response.SetStatusCode(500)
		} else {
			ctx.Response.AppendBody(result)
		}
	}
}

func register(etcd string, port int) {
	fmt.Printf("Registering to etcd %s\n", etcd)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcd},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create etcd client:", err)
		return
	}
	addresses, _ := net.InterfaceAddrs()
	ip := ""
	for _, addr := range addresses {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() && ipnet.IP.To4() != nil {
			ip = ipnet.IP.String()
		}
	}
	if ip == "" {
		fmt.Fprintln(os.Stderr, "Failed to get IP address")
		return
	}
	ipport := fmt.Sprintf("%s:%d", ip, port)
	lease := clientv3.NewLease(cli)
	grant, err := lease.Grant(context.TODO(), 10)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to grant lease from etcd:", err)
		return
	}
	_, err = cli.Put(context.TODO(), "dubbomesh/" + ipport, ipport, clientv3.WithLease(grant.ID))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to put to etcd:", err)
		return
	}
	fmt.Println("Register success:", ipport)
	go func() {
		for {
			time.Sleep(5 * time.Second)
			//fmt.Println("Heartbeat ...")
			_, err = lease.KeepAliveOnce(context.TODO(), grant.ID)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to send heartbeat:", err)
				return
			}
		}
	}()
}
