package dubbo

import (
	"io"
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

func (h Header) Write(w io.Writer) error {
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
	if _, err := w.Write([]byte{0xda, 0xbb, flags, h.Status}); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, h.RequestID); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, h.DataLength); err != nil {
		return err
	}
	return nil
}
