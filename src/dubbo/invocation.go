package dubbo

import "util"

type Invocation struct {
	DubboVersion     string
	ServiceName      string
	ServiceVersion   string
	MethodName       string
	MethodParamTypes string
	MethodArgs       string
	Attachments      map[string]string
}

func (inv Invocation) ToBytes() []byte {
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
	return data
}
