#!/usr/bin/env bash

set -euo pipefail

SPIN() {
  printf "⏳ %-45s" "$1"
}

PASS() {
  printf "\r✅ %-45s\n" "$1"
}

FAIL() {
  printf "\r❌ %-45s\n" "$1"
  exit 1
}


export HOME=/tmp/ssh-manager-test
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_DATA_HOME="$HOME/.local/share"
export XDG_CACHE_HOME="$HOME/.cache"

rm -rf "$HOME"
mkdir -p \
  "$XDG_CONFIG_HOME" \
  "$XDG_DATA_HOME" \
  "$XDG_CACHE_HOME"


SSH_HOST="${TEST_SSH_HOST:-"localhost"}"
SSH_PORT="${TEST_SSH_PORT:-22}"
SSH_USER="${TEST_SSH_USER:-"test"}"
SSH_PASS="${TEST_SSH_PASS:-"password"}"

# TEST 1
# -------------------------
SPIN "init creates dirs & config"
ssh-man-init >/tmp/init.log 2>&1 || FAIL "init"
[ -f "$XDG_CONFIG_HOME/ssh_man/config.yaml" ] || FAIL "app config yaml does note exist"
[ -d "$XDG_CONFIG_HOME/ssh_man/ssh" ] || FAIL "ssh config directory does not exist"
[ -d "$XDG_DATA_HOME/ssh_man/db/"] || FAIL "ssh database directory does not exist"
[ -d "$XDG_DATA_HOME/ssh_man/checksums/"] || FAIL "ssh checksum directory does not exist"
[ -d "$XDG_CONFIG_HOME/ssh_man/ssh/keystore" ] || FAIL "ssh keystore directory does not exist"
PASS "init creates dirs & config"


# TEST 2
# -------------------------
SPIN "flag parsing"
ssh-man --help >/tmp/help.log 2>&1 || FAIL "help"
grep -q Usage /tmp/help.log || FAIL "usage text"
PASS "flag parsing"


# TEST 3
# -------------------------
SPIN "register ssh host"
ssh-man -qa \
  --host dev-test \
  --hostname "$SSH_HOST" \
  --o "Port=$SSH_PORT" \
  --o "User=$SSH_USER" \
  >/tmp/add.log 2>&1 || FAIL "add-host"
PASS "register ssh host"


# Test 4
# -------------------------
SPIN "edit ssh host"
ssh-man -qe \
    --host dev-test \
    --o "User=newUser" \
    >/tmp/edit.log 2>&1 || FAIL "edit-host"
PASS "edit ssh host"

# Test 5
# -------------------------
SPIN "print ssh host"
ssh-man -gh \
    --host dev-test \
    > /tmp/print.log 2>&1 || FAIL "print-host"
grep -q "User newUser" /tmp/print.log || FAIL "print-host"
PASS "print ssh host"

# Test 6
# -------------------------
SPIN "delete ssh host"
ssh-man -qd \
    --host dev-test \
    > /tmp/delete.log 2>&1 || FAIL "delete-host"
PASS "delete-host"

# Test 7
# -------------------------
SPIN "sync config file to internal database"
ssh-man -qs \
        -f /test/ssh_config \
        > /tmp/sync.log 2>&1 || FAIL "sync-file-part1"
ssh-man -gh \
    --host dev-test \
    > /tmp/get_host_after_sync.log 2>&1 || FAIL "sync-file-part2"
PASS "sync-host"

# Test 8
# -------------------------
SPIN "test ssh using app generated config"
export SSHPASS="$TEST_SSH_PASS"
sshpass -e \
 ssh \
    -F "$XDG_CONFIG_HOME/ssh_man/ssh/config"
    -o BatchMode=no \
    -o StrictHostKeyChecking=no \
    -o UserKnownHostsFile=/dev/null \
    dev-test "echo HELLO WORLD" || FAIL "failed-SSH-config-check"
PASS "SSH-config-check"
echo "" 
echo "All test passed"