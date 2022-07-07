//go:build windows
// +build windows

package ztime

import "time"

func (f *fastime) now() time.Time {
	return time.Now().In(f.GetLocation())
}
