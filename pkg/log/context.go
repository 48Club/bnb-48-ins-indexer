package log

import (
	"sync"

	"go.uber.org/zap/zapcore"
)

var (
	Contexts sync.Map
)

type Context struct {
	RequestId string
}

func SetContext(requestId string) {
	ctx := Context{
		RequestId: "-",
	}
	ctx.RequestId = requestId
	Contexts.Store(Getgid(), ctx)
}

func GetContext() Context {
	if v, ok := Contexts.Load(Getgid()); ok {
		if ctx, ok := v.(Context); ok {
			return ctx
		}
	}
	return Context{RequestId: "-"}
}

func DelContext() {
	Contexts.Delete(Getgid())
}

func CallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	ctx := GetContext()

	enc.AppendString(ctx.RequestId)
	enc.AppendString(caller.TrimmedPath())
}
