package dubbo

type Invocation struct {
	DubboVersion     string
	ServiceName      string
	ServiceVersion   string
	MethodName       string
	MethodParamTypes string
	MethodArgs       string
	Attachments      map[string]string
}
