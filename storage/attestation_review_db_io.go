package storage

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

const reviewSchemaVersionMarker = "schema_version:1"

// DeleteBranchSessionsOptions controls optional safety behaviors for branch deletes.
// Zero-value options preserve existing behavior.
type DeleteBranchSessionsOptions struct {
	DryRun bool
	Logf   func(format string, args ...any)
}

// DeleteAllSessionsOptions controls optional safety behaviors for full-table deletes.
// Zero-value options preserve existing behavior.
type DeleteAllSessionsOptions struct {
	RequireConfirmation bool
	Confirmed           bool
	Logf                func(format string, args ...any)
}

// EnsureReviewDBDir creates the .git/lrc directory that stores reviews.db.
func EnsureReviewDBDir(lrcDir string) error {
	if err := MkdirAll(lrcDir, 0755); err != nil {
		return fmt.Errorf("failed to create review database directory %s: %w", lrcDir, err)
	}
	return nil
}

// OpenAttestationReviewDB opens the attestation review SQLite database with WAL and busy timeout.
func OpenAttestationReviewDB(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=%d", dbPath, sqliteBusyTimeoutMS())
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open review sqlite database %s: %w", dbPath, err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to connect review sqlite database %s: %w", dbPath, err)
	}
	return db, nil
}

// InitializeAttestationReviewSchema executes schema SQL for the review sessions table.
func InitializeAttestationReviewSchema(db *sql.DB, schema string) error {
	if db == nil {
		return fmt.Errorf("failed to initialize review schema: nil database handle")
	}
	if _, err := ExecSQL(db, schema); err != nil {
		return fmt.Errorf("failed to initialize review schema (%s): %w", compactSQL(schema), err)
	}
	// Optional schema marker check to keep schema evolution auditable without breaking existing schemas.
	if !strings.Contains(strings.ToLower(schema), reviewSchemaVersionMarker) {
		// Intentionally non-fatal for backward compatibility.
	}
	return nil
}

// InsertAttestationReviewSessionRow inserts a review session row for coverage tracking.
func InsertAttestationReviewSessionRow(db *sql.DB, treeHash, branch, action, timestamp, diffFilesJSON, reviewID string) error {
	if db == nil {
		return fmt.Errorf("failed to insert review session: nil database handle")
	}

	const insertSQL = `INSERT INTO review_sessions (tree_hash, branch, action, timestamp, diff_files, review_id)
	 VALUES (?, ?, ?, ?, ?, ?)`

	if _, err := ExecSQL(db, insertSQL, treeHash, branch, action, timestamp, diffFilesJSON, reviewID); err != nil {
		return fmt.Errorf("failed to insert review session row: %w", err)
	}
	return nil
}

// QueryAttestationReviewSessionCountByBranch returns total review sessions for one branch.
func QueryAttestationReviewSessionCountByBranch(db *sql.DB, branch string) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("failed to query review session count: nil database handle")
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM review_sessions WHERE branch = ?`, branch).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to query review session count for branch %q: %w", branch, err)
	}
	return count, nil
}

// QueryAttestationReviewedSessionsByBranch returns reviewed sessions in timestamp order.
func QueryAttestationReviewedSessionsByBranch(db *sql.DB, branch string) (*sql.Rows, error) {
	if db == nil {
		return nil, fmt.Errorf("failed to query reviewed sessions: nil database handle")
	}
	rows, err := db.Query(
		`SELECT id, tree_hash, branch, action, timestamp, diff_files, review_id
		 FROM review_sessions
		 WHERE branch = ? AND action = 'reviewed'
		 ORDER BY timestamp ASC`,
		branch,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviewed sessions for branch %q: %w", branch, err)
	}
	return rows, nil
}

// DeleteAttestationReviewSessionsByBranch deletes branch-local review sessions.
func DeleteAttestationReviewSessionsByBranch(db *sql.DB, branch string) (int64, error) {
	return DeleteAttestationReviewSessionsByBranchWithOptions(db, branch, DeleteBranchSessionsOptions{})
}

// DeleteAttestationReviewSessionsByBranchWithOptions deletes branch-local review sessions with optional dry-run/logging.
// DryRun=true returns the matching row count without mutating the database.
func DeleteAttestationReviewSessionsByBranchWithOptions(db *sql.DB, branch string, opts DeleteBranchSessionsOptions) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("failed to delete branch sessions: nil database handle")
	}

	if opts.DryRun {
		var count int64
		if err := db.QueryRow(`SELECT COUNT(*) FROM review_sessions WHERE branch = ?`, branch).Scan(&count); err != nil {
			return 0, fmt.Errorf("failed to dry-run review session delete for branch %q: %w", branch, err)
		}
		if opts.Logf != nil {
			opts.Logf("storage: dry-run delete for review_sessions branch=%q count=%d", branch, count)
		}
		return count, nil
	}

	result, err := ExecSQL(db, `DELETE FROM review_sessions WHERE branch = ?`, branch)
	if err != nil {
		return 0, fmt.Errorf("failed to delete review sessions for branch %q: %w", branch, err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to read branch delete rows-affected: %w", err)
	}
	if opts.Logf != nil {
		opts.Logf("storage: deleted review_sessions branch=%q affected=%d", branch, affected)
	}
	return affected, nil
}

// DeleteAllAttestationReviewSessions deletes all review sessions.
func DeleteAllAttestationReviewSessions(db *sql.DB) (int64, error) {
	return DeleteAllAttestationReviewSessionsWithOptions(db, DeleteAllSessionsOptions{})
}

// DeleteAllAttestationReviewSessionsWithOptions deletes all review sessions with optional confirmation/logging.
// Set RequireConfirmation=true and Confirmed=true to enforce caller confirmation without changing default behavior.
func DeleteAllAttestationReviewSessionsWithOptions(db *sql.DB, opts DeleteAllSessionsOptions) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("failed to delete all sessions: nil database handle")
	}
	if opts.RequireConfirmation && !opts.Confirmed {
		return 0, fmt.Errorf("failed to delete all review sessions: caller confirmation required")
	}

	result, err := ExecSQL(db, `DELETE FROM review_sessions`)
	if err != nil {
		return 0, fmt.Errorf("failed to delete all review sessions: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to read full delete rows-affected: %w", err)
	}
	if opts.Logf != nil {
		opts.Logf("storage: deleted all review_sessions affected=%d", affected)
	}
	return affected, nil
}

func compactSQL(query string) string {
	trimmedQuery := ""
	for _, line := range strings.Split(query, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if trimmedQuery != "" {
			trimmedQuery += " "
		}
		trimmedQuery += line
	}
	if len(trimmedQuery) > 240 {
		return trimmedQuery[:240] + "..."
	}
	return trimmedQuery
}
