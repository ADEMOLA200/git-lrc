package hooks

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteManagedHookScripts writes all lrc-managed hook scripts into dir.
func WriteManagedHookScripts(dir string, cfg TemplateConfig) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	scripts := map[string]string{
		"pre-commit":         GeneratePreCommitHook(cfg),
		"prepare-commit-msg": GeneratePrepareCommitMsgHook(cfg),
		"commit-msg":         GenerateCommitMsgHook(cfg),
		"post-commit":        GeneratePostCommitHook(cfg),
	}

	for name, content := range scripts {
		path := filepath.Join(dir, name)
		script := "#!/bin/sh\n" + content
		if err := os.WriteFile(path, []byte(script), 0755); err != nil {
			return fmt.Errorf("failed to write managed hook %s: %w", name, err)
		}
	}

	return nil
}
