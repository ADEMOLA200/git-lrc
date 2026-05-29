#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/installer-test-lib.sh"

trap cleanup_installer_test_env EXIT

require_claude_cli
setup_installer_test_env "lrc-installer-local"

bold ""
bold "══ Local Installer Bootstrap ═══════════════════════════════"

INSTALL_LOG="$TEST_LOG_DIR/install-local.log"
run_installer_capture \
	"$INSTALL_LOG" \
	env LRC_CLAUDE_PLUGIN_MARKETPLACE_SOURCE="$CLAUDE_LRC_DIR" \
	"$GIT_LRC_DIR/scripts/lrc-install.sh"
assert_exit_code "local installer exits successfully" "0" "$LAST_STATUS"

INSTALL_OUTPUT="$(cat "$INSTALL_LOG")"
assert_contains "installer defaults Claude-capable environments to git-only hooks" \
	"Claude detected; defaulting hook surface to git so the lrc plugin owns Claude integration" \
	"$INSTALL_OUTPUT"
assert_path_exists "installer writes lrc binary into temp home" "$TEST_HOME/.local/bin/lrc"
assert_ok "installed lrc binary reports a version" env HOME="$TEST_HOME" PATH="$TEST_RUNTIME_PATH" "$TEST_HOME/.local/bin/lrc" version

PLUGIN_LIST_JSON="$(wait_for_plugin_install "lrc@claude-lrc" || true)"
assert_contains "installer bootstraps lrc Claude plugin" '"id": "lrc@claude-lrc"' "$PLUGIN_LIST_JSON"
assert_path_missing "backend installer does not write legacy Claude skill" "$TEST_HOME/.claude/skills/lrc"
assert_path_missing "backend installer does not write legacy Claude hook assets" "$TEST_HOME/.lrc/claude"

finish_tests