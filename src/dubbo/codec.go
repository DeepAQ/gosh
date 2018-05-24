package dubbo

import (
	"io"
	"sync/atomic"
)

var reqId uint64

func writeRequest(w io.Writer, inv *Invocation) error {
	data := inv.ToBytes()
	h := Header{
		Req:           true,
		TwoWay:        true,
		Event:         false,
		Serialization: 6,
		Status:        0,
		RequestID:     atomic.AddUint64(&reqId, 1),
		DataLength:    uint32(len(data)),
	}
	if err := h.Write(w); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	return nil
}
