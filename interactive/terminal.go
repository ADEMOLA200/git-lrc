package interactive

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
)

var terminalOutputMu sync.Mutex

func SyncedPrintf(format string, args ...interface{}) {
	terminalOutputMu.Lock()
	defer terminalOutputMu.Unlock()
	fmt.Printf(format, args...)
	SyncFileSafely(os.Stdout)
}

func SyncedPrintln(args ...interface{}) {
	terminalOutputMu.Lock()
	defer terminalOutputMu.Unlock()
	fmt.Println(args...)
	SyncFileSafely(os.Stdout)
}

func SyncedFprintf(file *os.File, format string, args ...interface{}) {
	terminalOutputMu.Lock()
	defer terminalOutputMu.Unlock()
	fmt.Fprintf(file, format, args...)
	SyncFileSafely(file)
}

func SyncFileSafely(file *os.File) {
	if file == nil {
		return
	}

	if err := file.Sync(); err != nil {
		// Ignore common non-fatal sync errors for terminals and special files.
		if errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTSUP) {
			return
		}
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to sync output stream: %v\n", err)
	}
}
