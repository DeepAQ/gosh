package dubbo

import (
	"encoding/binary"
)

type Header struct {
	Req           bool
	TwoWay        bool
	Event         bool
	Serialization uint8
	Status        uint8
	RequestID     uint64
	DataLength    uint32
}

func (h Header) WriteTo(bytes []byte) {
	flags := h.Serialization
	if h.Event {
		flags += 1 << 5
	}
	if h.TwoWay {
		flags += 1 << 6
	}
	if h.Req {
		flags += 1 << 7
	}

	bytes[0] = 0xda
	bytes[1] = 0xbb
	bytes[2] = flags
	bytes[3] = h.Status
	binary.BigEndian.PutUint64(bytes[4:12], h.RequestID)
	binary.BigEndian.PutUint32(bytes[12:16], h.DataLength)
}
