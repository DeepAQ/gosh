package dubbo

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"reflect"
	"sync"
	"time"
)

var respMap sync.Map
var newReq chan []byte

type Response struct {
	Success bool
	Body    []byte
}

func Connect(remote string) error {
	conn, err := net.Dial("tcp", remote)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open connection to", remote, ":", err)
		return err
	}
	respMap = sync.Map{}
	newReq = make(chan []byte)
	// Writer
	go func() {
		for {
			req := <-newReq
			//fmt.Print("-> ")
			//for _, byte := range req {
			//	fmt.Printf("%02x ", byte)
			//}
			//fmt.Println()
			conn.Write(req)
		}
	}()
	// Reader
	go func() {
		header := make([]byte, 16)
		for {
			if _, err := conn.Read(header); err != nil {
				fmt.Fprintln(os.Stderr, "Failed to read header:", err)
				conn.Close()
				conn, _ = net.Dial("tcp", remote)
				continue
			}
			//fmt.Print("<- ")
			//for _, byte := range header {
			//	fmt.Printf("%02x ", byte)
			//}
			//fmt.Println()

			result := &Response{Success: true}

			bodyLen := int(binary.BigEndian.Uint32(header[12:]))
			var body []byte
			if bodyLen > 0 {
				body = make([]byte, bodyLen)
				read := 0
				for read < bodyLen {
					if i, err := conn.Read(body); err == nil {
						read += i
					} else {
						fmt.Fprintln(os.Stderr, "Failed to read body:", err)
						conn.Close()
						conn, _ = net.Dial("tcp", remote)
						continue
					}
				}
			}

			if header[0] != 0xda || header[1] != 0xbb {
				fmt.Fprintln(os.Stderr, "Bad magic:", header[0], header[1])
				result.Success = false
			}
			if header[3] != 20 {
				fmt.Fprintln(os.Stderr, "Server respond with status", header[3])
				result.Success = false
			}
			if bodyLen > 0 {
				var i, j int
				for i = 1; body[i] == '\r' || body[i] == '\n'; i++ {
				}
				for j = bodyLen - 1; body[j] == '\r' || body[j] == '\n'; j-- {
				}
				result.Body = body[i : j+1]
			}

			reqId := binary.BigEndian.Uint64(header[4:12])
			respChan, _ := respMap.Load(reqId)
			if respChan != nil {
				reflect.ValueOf(respChan).Send(reflect.ValueOf(&result))
			}
		}
	}()
	// Heartbeat
	hb := Header{
		Req:           true,
		TwoWay:        false,
		Event:         true,
		Serialization: 0,
		Status:        0,
		RequestID:     0,
		DataLength:    0,
	}
	go func() {
		for {
			time.Sleep(5 * time.Second)
			newReq <- hb.ToBytes()
		}
	}()
	fmt.Println("Dubbo: connected to", remote)
	return nil
}
