#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/installer-test-lib.sh"

trap cleanup_installer_test_env EXIT

require_claude_cli
setup_installer_test_env "lrc-plugin-bootstrap"
install_local_shell_installer_curl_shim

bold ""
bold "══ Plugin-First Bootstrap ══════════════════════════════════"

HOME="$TEST_HOME" PATH="$TEST_BASE_PATH" claude plugin marketplace add "$CLAUDE_LRC_DIR" >/dev/null
HOME="$TEST_HOME" PATH="$TEST_BASE_PATH" claude plugin install lrc@claude-lrc >/dev/null

PLUGIN_INSTALL_PATH="$(get_plugin_install_path "lrc@claude-lrc")"
ENSURE_PATH="$PLUGIN_INSTALL_PATH/scripts/ensure-lrc.sh"
assert_path_exists "plugin install path exists in temp Claude home" "$PLUGIN_INSTALL_PATH"
assert_path_exists "installed plugin exposes ensure-lrc.sh" "$ENSURE_PATH"
assert_path_missing "plugin-first flow starts without backend lrc installed" "$TEST_HOME/.local/bin/lrc"

FIRST_BOOTSTRAP_LOG="$TEST_LOG_DIR/plugin-first-bootstrap.log"
set +e
HOME="$TEST_HOME" PATH="$TEST_BASE_PATH" "$ENSURE_PATH" >"$FIRST_BOOTSTRAP_LOG" 2>&1
FIRST_BOOTSTRAP_STATUS=$?
set -e
assert_exit_code "plugin bootstrap installs backend successfully" "0" "$FIRST_BOOTSTRAP_STATUS"
assert_path_exists "plugin bootstrap installs lrc into temp home" "$TEST_HOME/.local/bin/lrc"
assert_ok "bootstrapped lrc binary reports a version" env HOME="$TEST_HOME" PATH="$TEST_RUNTIME_PATH" "$TEST_HOME/.local/bin/lrc" version

HOME="$TEST_HOME" PATH="$TEST_RUNTIME_PATH" "$TEST_HOME/.local/bin/lrc" hooks install --surface claude >/dev/null
assert_path_exists "legacy Claude skill can be seeded for migration test" "$TEST_HOME/.claude/skills/lrc"
assert_path_exists "legacy Claude hook assets can be seeded for migration test" "$TEST_HOME/.lrc/claude"

SECOND_BOOTSTRAP_LOG="$TEST_LOG_DIR/plugin-cleanup-bootstrap.log"
set +e
HOME="$TEST_HOME" PATH="$TEST_RUNTIME_PATH" "$ENSURE_PATH" >"$SECOND_BOOTSTRAP_LOG" 2>&1
SECOND_BOOTSTRAP_STATUS=$?
set -e
assert_exit_code "plugin rerun succeeds after seeding legacy Claude state" "0" "$SECOND_BOOTSTRAP_STATUS"
assert_path_missing "plugin rerun removes legacy Claude skill" "$TEST_HOME/.claude/skills/lrc"
assert_path_missing "plugin rerun removes legacy Claude hook assets" "$TEST_HOME/.lrc/claude"

finish_tests