package dubbo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

func Invoke(invocation *Invocation, conn net.Conn) ([]byte, error) {
	if err := writeRequest(conn, invocation); err != nil {
		fmt.Println("Failed to write:", err)
		return nil, err
	}
	var header [16]byte
	if _, err := conn.Read(header[:]); err != nil {
		fmt.Println("Failed to read header:", err)
		return nil, err
	}
	bodyLen := int(binary.BigEndian.Uint32(header[12:]))
	var body []byte
	if bodyLen > 0 {
		body = make([]byte, bodyLen)
		read := 0
		for read < bodyLen {
			if i, err := conn.Read(body); err == nil {
				read += i
			} else {
				fmt.Println("Failed to read body:", err)
				return nil, err
			}
		}
	}

	if header[3] != 20 {
		return nil, errors.New(fmt.Sprintf("Server respond with status %d", header[3]))
	}
	if bodyLen > 0 {
		var i, j int
		for i = 1; body[i] == '\r' || body[i] == '\n'; i++ {
		}
		for j = bodyLen - 1; body[j] == '\r' || body[j] == '\n'; j-- {
		}
		return body[i : j+1], nil
	}

	return []byte{}, nil
}
