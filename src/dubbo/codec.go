package dubbo

import (
	"io"
	"sync/atomic"
	"util"
)

var reqId uint64

func writeRequest(w io.Writer, inv *Invocation) error {
	data := append(util.ToJson(inv.DubboVersion), '\r', '\n')
	data = append(data, util.ToJson(inv.ServiceName)...)
	data = append(data, '\r', '\n')
	data = append(data, util.ToJson(inv.ServiceVersion)...)
	data = append(data, '\r', '\n')
	data = append(data, util.ToJson(inv.MethodName)...)
	data = append(data, '\r', '\n')
	data = append(data, util.ToJson(inv.MethodParamTypes)...)
	data = append(data, '\r', '\n')
	data = append(data, util.ToJson(inv.MethodArgs)...)
	data = append(data, '\r', '\n')
	data = append(data, util.ToJson(inv.Attachments)...)
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
