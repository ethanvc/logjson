package logjson

import (
	"fmt"
	"runtime"
)

func GetFilePathForLog(filePath string, line int) string {
	const keepPathCount = 2
	delimCount := 0
	result := filePath
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '/' || filePath[i] == '\\' {
			delimCount++
			if delimCount == keepPathCount {
				result = filePath[i+1:]
				break
			}
		}
	}
	return fmt.Sprintf("%s:%d", result, line)
}

func GetCallerFrame(pc uintptr) runtime.Frame {
	fs := runtime.CallersFrames([]uintptr{pc})
	frame, _ := fs.Next()
	return frame
}

func GetCaller(skip int) uintptr {
	var pcs [1]uintptr
	runtime.Callers(skip+2, pcs[:])
	return pcs[0]
}
