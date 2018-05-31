package dubbo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

var conn net.Conn
var respMap *sync.Map

var bufPool *sync.Pool
var respChanPool *sync.Pool

func Connect(remote string) error {
	var err error
	conn, err = net.Dial("tcp", remote)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open connection to", remote, ":", err)
		return err
	}
	// Create pools
	respMap = &sync.Map{}
	bufPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 16, 1024))
		},
	}
	respChanPool = &sync.Pool{
		New: func() interface{} {
			return make(chan []byte)
		},
	}

	// Reader
	go func() {
		header := make([]byte, 16)
		for {
			headerLen, err := conn.Read(header)
			if err != nil || headerLen < 16 {
				fmt.Println("Failed to read header:", err)
			}
			if header[3] != 20 {
				fmt.Println("Server respond with status:", header[3])
			}

			bodyLen := int(binary.BigEndian.Uint32(header[12:16]))
			var body []byte
			if bodyLen > 0 {
				body = make([]byte, bodyLen)
				read := 0
				for read < bodyLen {
					if i, err := conn.Read(body[read:]); err == nil {
						read += i
					} else {
						fmt.Println("Failed to read body:", err)
					}
				}
			}

			if bodyLen > 0 {
				var i, j int
				for i = 1; body[i] == '\r' || body[i] == '\n'; i++ {
				}
				for j = bodyLen - 1; body[j] == '\r' || body[j] == '\n'; j-- {
				}
				body = body[i : j+1]
			}

			reqId := binary.BigEndian.Uint64(header[4:12])
			respChan, _ := respMap.Load(reqId)
			if respChan != nil {
				respChan.(chan []byte) <- body
			}
		}
	}()

	// Heartbeat
	go func() {
		hbytes := make([]byte, 16)
		Header{
			Req:           true,
			TwoWay:        false,
			Event:         true,
			Serialization: 0,
			Status:        0,
			RequestID:     0,
			DataLength:    0,
		}.WriteTo(hbytes)
		for {
			time.Sleep(5 * time.Second)
			conn.Write(hbytes)
		}
	}()

	fmt.Println("Dubbo: connected to", remote)
	return nil
}
