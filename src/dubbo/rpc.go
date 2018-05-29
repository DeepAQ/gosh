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
	var resp [64]byte
	limit, err := conn.Read(resp[:])
	if err != nil {
		fmt.Println("Failed to read:", err)
		return nil, err
	}
	bodyLen := int(binary.BigEndian.Uint32(resp[12:]))
	body := resp[16:]
	if bodyLen > 0 && limit-16 < bodyLen {
		body = make([]byte, bodyLen)
		copy(body, resp[16:])
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

	if resp[3] != 20 {
		return nil, errors.New(fmt.Sprintf("Server respond with status %d", resp[3]))
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
