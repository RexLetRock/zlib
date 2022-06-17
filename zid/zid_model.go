package zid

import (
  "sync"
)

const (
	// PanicValue indicates when Next starts to panic
	PanicValue int64 = ((1 << 36) * 98 / 100) & ^1023
	// CriticalValue indicates when to renew the high 28 bits
	CriticalValue int64 = ((1 << 36) * 80 / 100) & ^1023
	// RenewIntervalMask indicates how often renew is performed if it fails
	RenewIntervalMask int64 = 0x20000000 - 1
)

// Option is for internal use only.
type Option func(*ZID)

// ZID is for internal use only.
type ZID struct {
	N     int64
	Step  int64
	Floor int64

	Name        string
	NoSec       bool
	Section     int8
	H28Verifier func(h28 int64) error

	sync.Mutex
	Renew func() error
}
