package consumer

import (
	"dubbo"
	"encoding/binary"
	"etcd"
	"fmt"
	"github.com/valyala/fasthttp"
	"math/rand"
	"strconv"
	"time"
	"unsafe"
	"util"
)

func Start(opts map[string]string) {
	port, _ := strconv.Atoi(opts["port"])
	if port <= 0 {
		port = 20000
	}
	fmt.Println("Starting consumer agent ...")

	hosts, weights, err := etcd.Query(opts["etcd"])
	if err != nil {
		fmt.Println("Failed to query from etcd:", err)
		return
	}
	totalServers = len(hosts)
	util.InitPools(8, 64, hosts)

	sumWeight := 0
	for _, weight := range weights {
		sumWeight += weight
	}

	serverProb = make([]float64, totalServers)
	for i := range serverProb {
		serverProb[i] = float64(weights[i]) / float64(sumWeight)
	}
	fmt.Println("Prob:", serverProb)
	rand.Seed(time.Now().UnixNano())
	// Load balancing method start
	//lbRT()
	// Load balancing method end

	// Listen
	fmt.Printf("Listening on port %d\n", port)
	if err := fasthttp.ListenAndServe(fmt.Sprintf(":%d", port), handler); err != nil {
		fmt.Println("Failed to listen:", err)
		return
	}
}

func handler(ctx *fasthttp.RequestCtx) {
	// Pick a server
	rnd := rand.Float64()
	sum := float64(0)
	var selected int
	prob := serverProb
	for selected = 0; rnd >= sum+prob[selected]; selected++ {
		sum += serverProb[selected]
	}

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

	buf := util.AcquireReqBuf()
	inv.WriteToBufAsCafe(buf)
	req := buf.Bytes()
	req[0] = 0xca
	req[1] = 0xfe
	req[2] = 0xbe
	req[3] = 0xef
	binary.BigEndian.PutUint32(req[4:8], uint32(len(req)-8))

	cw := util.AcquireConn(selected)
	//serverBegin := time.Now().UnixNano()
	if _, err := cw.Conn.Write(req); err != nil {
		cw.Conn.Close()
		cw = util.NewConn(selected)
		if cw.Conn == nil {
			fmt.Println("Failed to get conn")
			ctx.Response.SetStatusCode(500)
			return
		}
		if _, err := cw.Conn.Write(req); err != nil {
			fmt.Println("Failed to write req:", err)
			ctx.Response.SetStatusCode(500)
			cw.Conn.Close()
			return
		}
	}
	util.ReleaseReqBuf(8, buf)

	limit, err := cw.Conn.Read(cw.Buf)
	if err != nil {
		fmt.Println("Failed to read:", err)
		ctx.Response.SetStatusCode(500)
		cw.Conn.Close()
		return
	}
	if cw.Buf[0] != 0xca || cw.Buf[1] != 0xfe || limit < 8 {
		fmt.Println("Cafe bad magic")
		ctx.Response.SetStatusCode(500)
		cw.Conn.Close()
		return
	}

	if cw.Buf[2] == 0 && cw.Buf[3] == 0 {
		ctx.Response.SetStatusCode(500)
	}
	bodyLen := int(binary.BigEndian.Uint32(cw.Buf[4:8]))
	body := cw.Buf[8:limit]
	if bodyLen > 0 && limit-8 < bodyLen {
		body = make([]byte, bodyLen)
		copy(body, cw.Buf[8:limit])
		read := 0
		for read < bodyLen {
			if i, err := cw.Conn.Read(body[read:]); err == nil {
				read += i
			} else {
				fmt.Println("Failed to read body:", err)
				ctx.Response.SetStatusCode(500)
				cw.Conn.Close()
				return
			}
		}
	}
	//serverRT := time.Now().UnixNano() - serverBegin
	util.ReleaseConn(selected, cw)
	ctx.Response.SetBody(body)

	//if invokeRT != nil {
	//	atomic.AddInt64(&invokeRT[selected], serverRT/1E3)
	//}
	//if invokeCount != nil {
	//	atomic.AddUint32(&invokeCount[selected], 1)
	//}
}
