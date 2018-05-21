package dubbo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
)

func Invoke(invocation *Invocation, conn net.Conn) ([]byte, error) {
	if err := writeRequest(conn, invocation); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to write:", err)
		return nil, err
	}
	header := make([]byte, 16)
	if _, err := conn.Read(header); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read header:", err)
		return nil, err
	}
	if header[3] != 20 {
		return nil, errors.New(fmt.Sprintf("Server respond with status %d", header[3]))
	}
	bodyLen := binary.BigEndian.Uint32(header[12:])
	body := make([]byte, bodyLen)
	if _, err := conn.Read(body); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read body:", err)
		return nil, err
	}
	var i, j uint32
	for i = 1; body[i] == '\r' || body[i] == '\n'; i++ {
	}
	for j = bodyLen - 1; body[j] == '\r' || body[j] == '\n'; j-- {
	}
	return body[i : j+1], nil
}
