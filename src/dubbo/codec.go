package dubbo

import (
	"io"
	"encoding/json"
	"sync/atomic"
)

var reqId uint64

func writeRequest(w io.Writer, inv *Invocation) error {
	data := append(toJson(inv.DubboVersion), '\r', '\n')
	data = append(data, toJson(inv.ServiceName)...)
	data = append(data, '\r', '\n')
	data = append(data, toJson(inv.ServiceVersion)...)
	data = append(data, '\r', '\n')
	data = append(data, toJson(inv.MethodName)...)
	data = append(data, '\r', '\n')
	data = append(data, toJson(inv.MethodParamTypes)...)
	data = append(data, '\r', '\n')
	data = append(data, toJson(inv.MethodArgs)...)
	data = append(data, '\r', '\n')
	data = append(data, toJson(inv.Attachments)...)
	data = append(data, '\r', '\n')
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

func toJson(v interface{}) []byte {
	result, _ := json.Marshal(v)
	return result
}
