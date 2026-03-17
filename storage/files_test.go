package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChmodSecretFileMode(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "secret.txt")
	if err := os.WriteFile(path, []byte("secret"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	if err := Chmod(path, 0600); err != nil {
		t.Fatalf("failed to chmod file: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("unexpected mode: got %o want %o", got, os.FileMode(0600))
	}
}

func TestChmodExecutableMode(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "tool.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho ok\n"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	if err := Chmod(path, 0755); err != nil {
		t.Fatalf("failed to chmod file: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0755 {
		t.Fatalf("unexpected mode: got %o want %o", got, os.FileMode(0755))
	}
}
