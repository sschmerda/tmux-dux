#!/usr/bin/env bash
set -euo pipefail

CURRENT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="$CURRENT_DIR/bin/tmux-dux"
INSTALL_SCRIPT="$CURRENT_DIR/scripts/install.sh"

if [ ! -x "$BINARY" ] && [ -x "$INSTALL_SCRIPT" ]; then
  TMUX_DUX_INSTALL_DIR="$CURRENT_DIR/bin" "$INSTALL_SCRIPT" >/dev/null 2>&1 || true
fi

if [ -x "$BINARY" ]; then
  tmux set-environment -g TMUX_DUX_BIN "$BINARY"
else
  tmux set-environment -g TMUX_DUX_BIN "tmux-dux"
fi

key="$(tmux show-option -gqv @tmux-dux-key)"
if [ -n "$key" ]; then
  tmux bind-key "$key" run-shell '"${TMUX_DUX_BIN:-tmux-dux}" popup'
fi
