package dubbo

import (
	"sync/atomic"
)

var reqId uint64

func Invoke(inv *Invocation) []byte {
	invBytes := inv.ToBytes()
	newReqId := atomic.AddUint64(&reqId, 1)
	header := Header{
		Req:           true,
		TwoWay:        true,
		Event:         false,
		Serialization: 6,
		Status:        0,
		RequestID:     newReqId,
		DataLength:    uint32(len(invBytes)),
	}
	respChan := make(chan []byte)
	respMap.Store(newReqId, respChan)
	newReq[int(newReqId)%len(newReq)] <- append(header.ToBytes(), invBytes...)
	resp := <-respChan
	close(respChan)
	respMap.Delete(newReqId)
	return resp
}
