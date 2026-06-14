package lrcrules

import (
	"os"
	"path/filepath"

	"github.com/HexmosTech/git-lrc/storage"
)

const rootReadmeContent = `# .lrc/ — Repository Rules

This directory teaches the LiveReview AI reviewer about this repository:
its conventions, what's intentionally off-limits, and which files it
shouldn't review at all.

- rules/    — markdown guidance sent to the reviewer (see rules/README.md)
- ignore    — gitignore-style patterns for files/paths to exclude from review
- policy/   — machine-readable settings consumed directly by git-lrc
              (not yet enforced)

Run ` + "`lrc config check`" + ` to validate this directory and
` + "`lrc config preview`" + ` to see exactly what will be sent to the
reviewer.
`

const rulesReadmeContent = `# rules/ — Reviewer Guidance

LiveReview concatenates every other *.md file in this directory (this
README is excluded) in lexicographic order into a single instruction
bundle, with a "## rules/<file>.md" header above each file's content.
Empty files are skipped.

The combined bundle is limited to ` + "`lrcrules.CharLimit`" + ` (3000)
characters. Run ` + "`lrc config check`" + ` to verify you're within the
limit and ` + "`lrc config preview`" + ` to see the exact bundle.

Keep it short: capture the handful of ideas that repeatedly affect review
decisions (e.g. "prefer direct SQL over ORM abstractions", "avoid new
infrastructure dependencies").
`

const ignoreContent = `# .lrc/ignore — gitignore-style patterns
#
# Paths are matched relative to the repository root, using the same syntax
# as .gitignore (comments, blank lines, "**", negation with "!", etc.).
# Files matching a pattern here are excluded from AI review.
`

// scaffoldFile describes one file Init may create.
type scaffoldFile struct {
	relPath string // relative to .lrc/
	content string
}

func scaffoldFiles() []scaffoldFile {
	return []scaffoldFile{
		{"README.md", rootReadmeContent},
		{"ignore", ignoreContent},
		{"rules/README.md", rulesReadmeContent},
		{"rules/design.md", ""},
		{"rules/security.md", ""},
		{"rules/style.md", ""},
		{"policy/tools.toml", ""},
	}
}

// Init scaffolds .lrc/ under repoRoot idempotently: existing files and
// directories are left untouched. Returns the list of paths (relative to
// repoRoot, using "/" separators) that were created.
func Init(repoRoot string) ([]string, error) {
	lrcDir := filepath.Join(repoRoot, ".lrc")
	var created []string

	for _, f := range scaffoldFiles() {
		fullPath := filepath.Join(lrcDir, f.relPath)
		if _, err := os.Stat(fullPath); err == nil {
			continue
		} else if !os.IsNotExist(err) {
			return created, err
		}

		if err := storage.WriteFileAtomically(fullPath, []byte(f.content), 0o644); err != nil {
			return created, err
		}
		created = append(created, filepath.ToSlash(filepath.Join(".lrc", f.relPath)))
	}

	return created, nil
}
