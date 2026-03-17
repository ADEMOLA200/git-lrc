package interactive

import (
	"fmt"
	"os"
	"sync"
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
		// Platform-specific filtering keeps existing non-Windows behavior;
		// on Windows it ignores expected terminal sync errors (for example ERROR_INVALID_HANDLE).
		if isIgnorableSyncError(err) {
			return
		}
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to sync output stream: %v\n", err)
	}
}
