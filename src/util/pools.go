package util

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

var reqBufPool *sync.Pool
var reqConnPool *sync.Pool

// NewBufferPool creates a new BufferPool bounded to the given size.
func InitPools(remote string) {
	reqBufPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 16, 1024))
		},
	}
	reqConnPool = &sync.Pool{
		New: func() interface{} {
			conn, err := net.Dial("tcp", remote)
			if err != nil {
				fmt.Println("Failed to create new conn:", err)
			}
			return conn
		},
	}
}

func AcquireReqBuf() *bytes.Buffer {
	return reqBufPool.Get().(*bytes.Buffer)
}

func ReleaseReqBuf(buf *bytes.Buffer) {
	buf.Truncate(16)
	reqBufPool.Put(buf)
}

func AcquireConn() net.Conn {
	conn := reqConnPool.Get()
	if conn != nil {
		return conn.(net.Conn)
	} else {
		return NewConn()
	}
}

func NewConn() net.Conn {
	conn := reqConnPool.New()
	if conn != nil {
		return conn.(net.Conn)
	} else {
		return nil
	}
}

func ReleaseConn(conn net.Conn) {
	reqConnPool.Put(conn)
}
