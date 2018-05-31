package util

import (
	"sync"
)

var cafeRespBytesPool *sync.Pool

func InitCafePools(respBytesLen int) {
	cafeRespBytesPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, respBytesLen)
		},
	}
}

func AcquireCafeRespBytes() []byte {
	return cafeRespBytesPool.Get().([]byte)
}

func ReleaseCafeRespBytes(bytes []byte) {
	cafeRespBytesPool.Put(bytes)
}
