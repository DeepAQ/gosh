package dubbo

import (
	"bytes"
	"sync/atomic"
)

var globalReqId uint64

func Invoke(inv Invocation) ([]byte, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	inv.WriteToBuf(buf)
	data := buf.Bytes()
	newReqId := atomic.AddUint64(&globalReqId, 1)
	Header{
		Req:           true,
		TwoWay:        true,
		Event:         false,
		Serialization: 6,
		Status:        0,
		RequestID:     newReqId,
		DataLength:    uint32(len(data) - 16),
	}.WriteTo(data)

	respChan := respChanPool.Get()
	respMap.Store(newReqId, respChan)
	_, err := conn.Write(data)

	//fmt.Print("-> ")
	//for _, byte := range data {
	//	fmt.Printf("%02x ", byte)
	//}
	//fmt.Println("\n")

	buf.Truncate(16)
	bufPool.Put(buf)
	if err != nil {
		respChanPool.Put(respChan)
		respMap.Delete(newReqId)
		return nil, err
	}

	resp := <-respChan.(chan []byte)
	respChanPool.Put(respChan)
	respMap.Delete(newReqId)
	return resp, nil
}
