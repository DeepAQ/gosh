package dubbo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"util"
)

var globalReqId uint64

func Invoke(inv Invocation) ([]byte, error) {
	conn := util.AcquireConn()
	if err := writeRequest(conn, inv); err != nil {
		conn.Close()
		conn = util.NewConn()
		if conn == nil {
			fmt.Println("Failed to get conn")
			return nil, err
		}
		if err := writeRequest(conn, inv); err != nil {
			conn.Close()
			fmt.Println("Failed to write req:", err)
			return nil, err
		}
	}
	var resp [64]byte
	limit, err := conn.Read(resp[:])
	if err != nil {
		fmt.Println("Failed to read:", err)
		return nil, err
	}
	bodyLen := int(binary.BigEndian.Uint32(resp[12:]))
	body := resp[16:]
	if bodyLen > 0 && limit-16 < bodyLen {
		body = make([]byte, bodyLen)
		copy(body, resp[16:])
		read := 0
		for read < bodyLen {
			if i, err := conn.Read(body); err == nil {
				read += i
			} else {
				fmt.Println("Failed to read body:", err)
				return nil, err
			}
		}
	}
	util.ReleaseConn(conn)

	if resp[3] != 20 {
		return nil, errors.New(fmt.Sprintf("Server respond with status %d", resp[3]))
	}
	if bodyLen > 0 {
		var i, j int
		for i = 1; body[i] == '\r' || body[i] == '\n'; i++ {
		}
		for j = bodyLen - 1; body[j] == '\r' || body[j] == '\n'; j-- {
		}
		return body[i : j+1], nil
	}

	return []byte{}, nil
}

func writeRequest(w io.Writer, inv Invocation) error {
	buf := util.AcquireReqBuf()
	inv.WriteToBuf(buf)
	data := buf.Bytes()
	h := Header{
		Req:           true,
		TwoWay:        true,
		Event:         false,
		Serialization: 6,
		Status:        0,
		RequestID:     atomic.AddUint64(&globalReqId, 1),
		DataLength:    uint32(len(data) - 16),
	}
	h.WriteTo(data)
	_, err := w.Write(data)
	util.ReleaseReqBuf(buf)
	if err != nil {
		return err
	}
	return nil
}
