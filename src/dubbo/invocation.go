package dubbo

import (
	"bytes"
	"fmt"
	"util"
)

type Invocation struct {
	DubboVersion     string
	ServiceName      []byte
	ServiceVersion   []byte
	MethodName       []byte
	MethodParamTypes []byte
	MethodArgs       []byte
	Attachments      map[string]string
}

func (inv Invocation) ToBytes() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('"')
	buf.WriteString(inv.DubboVersion)
	buf.WriteByte('"')
	buf.WriteByte('\n')

	buf.WriteByte('"')
	buf.Write(inv.ServiceName)
	buf.WriteByte('"')
	buf.WriteByte('\n')

	if inv.ServiceVersion != nil {
		buf.WriteByte('"')
		buf.Write(inv.ServiceVersion)
		buf.WriteByte('"')
		buf.WriteByte('\n')
	} else {
		buf.WriteString("\"\"\n")
	}

	buf.WriteByte('"')
	buf.Write(inv.MethodName)
	buf.WriteByte('"')
	buf.WriteByte('\n')

	buf.WriteByte('"')
	buf.Write(inv.MethodParamTypes)
	buf.WriteByte('"')
	buf.WriteByte('\n')

	buf.WriteByte('"')
	buf.Write(inv.MethodArgs)
	buf.WriteByte('"')
	buf.WriteByte('\n')

	if inv.Attachments != nil {
		buf.Write(util.ToJson(inv.Attachments))
		buf.WriteByte('\n')
	} else {
		buf.WriteString("null\n")
	}
	fmt.Println(string(buf.Bytes()))
	return buf.Bytes()
}
