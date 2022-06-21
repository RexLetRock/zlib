// Get Goroutine ID Superfast

package zgoid

import (
	"bytes"
	"runtime"
	"strconv"
)

func ExtractGID(s []byte) int64 {
	s = s[len("goroutine "):]
	s = s[:bytes.IndexByte(s, ' ')]
	gid, _ := strconv.ParseInt(string(s), 10, 64)
	return gid
}

// Parse the goid from runtime.Stack() output. Slow, but it works.
func getSlow() int64 {
	var buf [64]byte
	return ExtractGID(buf[:runtime.Stack(buf[:], false)])
}
