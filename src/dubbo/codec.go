package dubbo

import (
	"io"
	"sync/atomic"
	"util"
)

var reqId uint64

func writeRequest(w io.Writer, inv *Invocation) error {
	buf := util.BufPool.Get()
	inv.WriteToBuf(buf)
	data := buf.Bytes()
	h := Header{
		Req:           true,
		TwoWay:        true,
		Event:         false,
		Serialization: 6,
		Status:        0,
		RequestID:     atomic.AddUint64(&reqId, 1),
		DataLength:    uint32(len(data) - 16),
	}
	h.WriteTo(data)
	_, err := w.Write(data)
	util.BufPool.Put(buf)
	if err != nil {
		return err
	}
	return nil
}
