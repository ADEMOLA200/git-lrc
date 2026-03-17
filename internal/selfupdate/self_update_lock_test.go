package selfupdate

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/HexmosTech/git-lrc/storage"
)

func TestSelfUpdateLockHelperProcess(t *testing.T) {
	if os.Getenv("LRC_TEST_HOLD_LOCK_HELPER") != "1" {
		return
	}

	release, acquired, err := acquireUpdateLock(false, "test-helper", false)
	if err != nil || !acquired {
		os.Exit(2)
	}
	defer release()

	readyFile := os.Getenv("LRC_TEST_READY_FILE")
	if readyFile != "" {
		_ = os.WriteFile(readyFile, []byte("ready"), 0644)
	}

	time.Sleep(20 * time.Second)
	os.Exit(0)
}

func TestAcquireUpdateLockForceRecoversLiveHolder(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("force recovery process test is unix-only")
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	readyFile := filepath.Join(tmpHome, "helper-ready")
	helper := exec.Command(os.Args[0], "-test.run=TestSelfUpdateLockHelperProcess")
	helper.Env = append(os.Environ(),
		"LRC_TEST_HOLD_LOCK_HELPER=1",
		"LRC_TEST_READY_FILE="+readyFile,
		"HOME="+tmpHome,
		"USERPROFILE="+tmpHome,
	)
	if err := helper.Start(); err != nil {
		t.Fatalf("failed to start helper process: %v", err)
	}
	defer func() {
		if helper.Process != nil {
			_ = helper.Process.Kill()
		}
	}()

	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(readyFile); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("helper process did not signal ready in time")
		}
		time.Sleep(50 * time.Millisecond)
	}

	release, acquired, err := acquireUpdateLock(false, "test-main", false)
	if err != nil {
		t.Fatalf("unexpected error acquiring lock without force: %v", err)
	}
	if acquired {
		release()
		t.Fatal("expected lock to be held by helper process")
	}

	release, acquired, err = acquireUpdateLock(true, "test-main-force", false)
	if err != nil {
		t.Fatalf("force lock recovery failed: %v", err)
	}
	if !acquired {
		t.Fatal("expected force lock recovery to acquire lock")
	}
	release()

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- helper.Wait()
	}()

	select {
	case <-time.After(5 * time.Second):
		t.Fatal("helper process did not exit after force recovery")
	case <-waitCh:
		// Exiting due to signal is expected after force recovery.
	}
}

func TestSavePendingUpdateStateAtomicWrite(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	state := &storage.PendingUpdateState{
		Version:          "v0.1.99",
		StagedBinaryPath: "/tmp/lrc-test-bin",
		DownloadedAt:     "2026-03-08T00:00:00Z",
	}

	if err := savePendingUpdateState(state); err != nil {
		t.Fatalf("savePendingUpdateState failed: %v", err)
	}

	statePath, err := pendingUpdateStatePath()
	if err != nil {
		t.Fatalf("pendingUpdateStatePath failed: %v", err)
	}
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("failed to read pending update state file: %v", err)
	}

	var parsed storage.PendingUpdateState
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("saved state is not valid JSON: %v", err)
	}
	if parsed.Version != state.Version || parsed.StagedBinaryPath != state.StagedBinaryPath {
		t.Fatalf("saved state mismatch: got %+v want %+v", parsed, *state)
	}
	if parsed.IntegrityHash == "" {
		t.Fatal("expected integrity hash to be set")
	}
}

func TestLoadPendingUpdateStateAcceptsLegacyStateWithoutIntegrityHash(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	statePath, err := pendingUpdateStatePath()
	if err != nil {
		t.Fatalf("pendingUpdateStatePath failed: %v", err)
	}
	legacy := `{
  "version": "v0.2.0",
  "staged_binary_path": "/tmp/lrc-bin",
  "downloaded_at": "2026-03-08T00:00:00Z"
}`
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		t.Fatalf("mkdir state dir failed: %v", err)
	}
	if err := os.WriteFile(statePath, []byte(legacy), 0644); err != nil {
		t.Fatalf("write legacy state failed: %v", err)
	}

	state, err := loadPendingUpdateState()
	if err != nil {
		t.Fatalf("loadPendingUpdateState failed for legacy state: %v", err)
	}
	if state == nil || state.Version != "v0.2.0" {
		t.Fatalf("unexpected state: %+v", state)
	}
}

func TestLoadPendingUpdateStateDetectsTampering(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	state := &storage.PendingUpdateState{
		Version:          "v0.3.0",
		StagedBinaryPath: "/tmp/lrc-test-bin",
		DownloadedAt:     "2026-03-08T00:00:00Z",
	}
	if err := savePendingUpdateState(state); err != nil {
		t.Fatalf("savePendingUpdateState failed: %v", err)
	}

	statePath, err := pendingUpdateStatePath()
	if err != nil {
		t.Fatalf("pendingUpdateStatePath failed: %v", err)
	}
	mutated := `{
  "version": "v0.3.0",
  "staged_binary_path": "/tmp/tampered-binary",
  "downloaded_at": "2026-03-08T00:00:00Z",
  "integrity_hash": "` + state.IntegrityHash + `"
}`
	if err := os.WriteFile(statePath, []byte(mutated), 0644); err != nil {
		t.Fatalf("failed to write tampered state: %v", err)
	}

	if _, err := loadPendingUpdateState(); err == nil {
		t.Fatal("expected tampered state integrity failure")
	}
}

func TestLoadPendingUpdateStateAcceptsIntegrityHashWithWhitespace(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	state := &storage.PendingUpdateState{
		Version:          "v0.4.0",
		StagedBinaryPath: "/tmp/lrc-test-bin",
		DownloadedAt:     "2026-03-08T00:00:00Z",
	}
	if err := savePendingUpdateState(state); err != nil {
		t.Fatalf("savePendingUpdateState failed: %v", err)
	}

	statePath, err := pendingUpdateStatePath()
	if err != nil {
		t.Fatalf("pendingUpdateStatePath failed: %v", err)
	}
	mutated := `{
  "version": "v0.4.0",
  "staged_binary_path": "/tmp/lrc-test-bin",
  "downloaded_at": "2026-03-08T00:00:00Z",
  "integrity_hash": "  ` + state.IntegrityHash + `  "
}`
	if err := os.WriteFile(statePath, []byte(mutated), 0644); err != nil {
		t.Fatalf("failed to write state: %v", err)
	}

	loaded, err := loadPendingUpdateState()
	if err != nil {
		t.Fatalf("expected whitespace-tolerant integrity hash: %v", err)
	}
	if loaded == nil || loaded.Version != "v0.4.0" {
		t.Fatalf("unexpected loaded state: %+v", loaded)
	}
}

func TestLoadPendingUpdateStateAcceptsLegacyHashFormat(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)

	legacyState := &storage.PendingUpdateState{
		Version:          "v0.5.0",
		StagedBinaryPath: "/tmp/lrc-legacy-bin",
		DownloadedAt:     "2026-03-08T00:00:00Z",
	}
	legacyHash := pendingUpdateStateIntegrityHashLegacy(legacyState)

	statePath, err := pendingUpdateStatePath()
	if err != nil {
		t.Fatalf("pendingUpdateStatePath failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		t.Fatalf("mkdir state dir failed: %v", err)
	}
	body := `{
  "version": "v0.5.0",
  "staged_binary_path": "/tmp/lrc-legacy-bin",
  "downloaded_at": "2026-03-08T00:00:00Z",
  "integrity_hash": "` + legacyHash + `"
}`
	if err := os.WriteFile(statePath, []byte(body), 0644); err != nil {
		t.Fatalf("write legacy-hash state failed: %v", err)
	}

	loaded, err := loadPendingUpdateState()
	if err != nil {
		t.Fatalf("expected legacy hash compatibility: %v", err)
	}
	if loaded == nil || loaded.Version != "v0.5.0" {
		t.Fatalf("unexpected loaded state: %+v", loaded)
	}
}
