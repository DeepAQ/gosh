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

func (h Header) ToBytes() []byte {
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
	bytes := []byte{0xda, 0xbb, flags, h.Status}

	reqIdBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(reqIdBytes, h.RequestID)
	bytes = append(bytes, reqIdBytes...)

	dataLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(dataLenBytes, h.DataLength)
	bytes = append(bytes, dataLenBytes...)

	return bytes
}
