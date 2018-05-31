package util

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

var reqBufPool *sync.Pool
var reqConnPool []*sync.Pool

type ConnWrapper struct {
	Conn net.Conn
	Buf  []byte
}

func InitPools(headerLen int, connBufLen int, remotes []string) {
	reqBufPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, headerLen, 1024))
		},
	}
	reqConnPool = make([]*sync.Pool, len(remotes))
	for i, remote := range remotes {
		reqConnPool[i] = &sync.Pool{
			New: func() interface{} {
				conn, err := net.Dial("tcp", remote)
				if err != nil {
					fmt.Println("Failed to create new conn:", err)
				}
				return &ConnWrapper{
					Conn: conn.(net.Conn),
					Buf:  make([]byte, connBufLen),
				}
			},
		}
	}
}

func AcquireReqBuf() *bytes.Buffer {
	return reqBufPool.Get().(*bytes.Buffer)
}

func ReleaseReqBuf(headerLen int, buf *bytes.Buffer) {
	buf.Truncate(headerLen)
	reqBufPool.Put(buf)
}

func AcquireConn(index int) *ConnWrapper {
	conn := reqConnPool[index].Get()
	if conn != nil {
		return conn.(*ConnWrapper)
	} else {
		return NewConn(index)
	}
}

func NewConn(index int) *ConnWrapper {
	conn := reqConnPool[index].New()
	if conn != nil {
		return conn.(*ConnWrapper)
	} else {
		return nil
	}
}

func ReleaseConn(index int, cw *ConnWrapper) {
	reqConnPool[index].Put(cw)
}
