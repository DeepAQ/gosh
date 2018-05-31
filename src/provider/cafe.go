package provider

import (
	"bytes"
	"dubbo"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"util"
)

func listenCafe(port int) {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: port,
	})
	if err != nil {
		fmt.Println("Failed to listen:", err)
		return
	}
	util.InitCafePools(64)
	for {
		conn, err := listener.AcceptTCP()
		conn.SetNoDelay(true)
		if err != nil {
			fmt.Println("Failed to accept:", err)
		}
		go cafeHandler(conn)
	}
}

func cafeHandler(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		limit, err := conn.Read(buf)
		if err != nil {
			conn.Close()
			return
		}
		if buf[0] != 0xca || buf[1] != 0xfe || limit < 8 {
			fmt.Println("Cafe bad magic")
			conn.Close()
			return
		}
		bodyLen := int(binary.BigEndian.Uint32(buf[4:8]))
		body := buf[8:limit]
		if limit-8 < bodyLen {
			body = make([]byte, bodyLen)
			copy(body, buf[8:limit])
			read := limit - 8
			for read < bodyLen {
				i, err := conn.Read(buf)
				if err != nil {
					fmt.Println("Failed to read from client:", err)
					conn.Close()
					return
				}
				read += i
			}
		}
		s1 := bytes.IndexByte(body, 0xff)
		s2 := bytes.IndexByte(body[s1+1:], 0xff)
		s3 := bytes.IndexByte(body[s1+s2+2:], 0xff)
		if s1 < 0 || s2 < 0 || s3 < 0 {
			fmt.Println("Corrupted body")
			cafeWriteError(conn)
			continue
		}
		inv := dubbo.Invocation{
			DubboVersion:     "2.0.0",
			ServiceName:      body[0:s1],
			MethodName:       body[s1+1 : s1+1+s2],
			MethodParamTypes: body[s1+s2+2 : s1+s2+2+s3],
			MethodArgs:       body[s1+s2+s3+3:],
		}
		result, err := dubbo.Invoke(inv)
		if err != nil || result == nil {
			fmt.Println("Invocation error:", err)
			cafeWriteError(conn)
		} else {
			resp := util.AcquireCafeRespBytes()
			resp[0] = 0xca
			resp[1] = 0xfe
			resp[2] = 0xbe
			resp[3] = 0xef
			bodyLen := len(result)
			binary.BigEndian.PutUint32(resp[4:8], uint32(bodyLen))
			if bodyLen <= 64-8 {
				copy(resp[8:], result)
				conn.Write(resp[0 : 8+bodyLen])
			} else {
				conn.Write(resp[0:8])
				conn.Write(result)
			}
			util.ReleaseCafeRespBytes(resp)
		}
	}
}

func cafeWriteError(w io.Writer) {
	resp := util.AcquireCafeRespBytes()
	resp[0] = 0xca
	resp[1] = 0xfe
	resp[2] = 0x00
	resp[3] = 0x00
	resp[4] = 0x00
	resp[5] = 0x00
	resp[6] = 0x00
	resp[7] = 0x00
	w.Write(resp[0:8])
	util.ReleaseCafeRespBytes(resp)
}
