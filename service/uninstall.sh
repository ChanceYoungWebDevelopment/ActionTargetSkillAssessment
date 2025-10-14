#!/usr/bin/env bash
set -euo pipefail

APP=at-ping
BIN_DST="/usr/local/bin/$APP"
UNIT_DST="/etc/systemd/system/$APP.service"

echo "==> Stopping and disabling $APP"
sudo systemctl disable --now "$APP.service" || true
sudo rm -f "$UNIT_DST"
sudo systemctl daemon-reload || true

echo "==> Removing binary (leaving /etc/$APP in place)"
sudo rm -f "$BIN_DST"

echo "==> Done. (If desired: sudo rm -rf /etc/$APP)"
