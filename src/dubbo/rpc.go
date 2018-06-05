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
	util.AcquireSem()
	cw := util.AcquireConn(0)
	if err := writeRequest(cw.Conn, inv); err != nil {
		cw.Conn.Close()
		cw = util.NewConn(0)
		if cw.Conn == nil {
			fmt.Println("Failed to get conn")
			util.ReleaseSem()
			return nil, err
		}
		if err := writeRequest(cw.Conn, inv); err != nil {
			fmt.Println("Failed to write req:", err)
			util.ReleaseSem()
			cw.Conn.Close()
			return nil, err
		}
	}
	limit, err := cw.Conn.Read(cw.Buf)
	if err != nil {
		fmt.Println("Failed to read:", err)
		util.ReleaseSem()
		cw.Conn.Close()
		return nil, err
	}
	bodyLen := int(binary.BigEndian.Uint32(cw.Buf[12:16]))
	body := cw.Buf[16:limit]
	if bodyLen > 0 && limit-16 < bodyLen {
		body = make([]byte, bodyLen)
		copy(body, cw.Buf[16:limit])
		read := 0
		for read < bodyLen {
			if i, err := cw.Conn.Read(body[read:]); err == nil {
				read += i
			} else {
				fmt.Println("Failed to read body:", err)
				util.ReleaseSem()
				cw.Conn.Close()
				return nil, err
			}
		}
	}
	util.ReleaseConn(0, cw)
	util.ReleaseSem()

	if cw.Buf[3] != 20 {
		return nil, errors.New(fmt.Sprintf("Server respond with status %d", cw.Buf[3]))
	}

	if bodyLen > 0 {
		var i, j int
		for i = 1; body[i] == '\r' || body[i] == '\n'; i++ {
		}
		for j = bodyLen - 1; body[j] == '\r' || body[j] == '\n'; j-- {
		}
		return body[i : j+1], nil
	}

	return nil, nil
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
	util.ReleaseReqBuf(16, buf)
	if err != nil {
		return err
	}
	return nil
}
