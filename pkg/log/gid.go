package log

import (
	"bytes"
	"runtime"
	"strconv"
)

// Getgid get current goroutine id
func Getgid() int64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	i, _ := strconv.ParseInt(string(b), 10, 64)
	return i
}
