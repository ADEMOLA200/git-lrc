//go:build !windows

package interactive

import (
	"errors"
	"syscall"
)

func isIgnorableSyncError(err error) bool {
	// Keep this list narrow: other errors should be surfaced by SyncFileSafely.
	return errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTSUP)
}
