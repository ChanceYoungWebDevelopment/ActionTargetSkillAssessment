#!/usr/bin/env bash
set -euo pipefail

APP=at-ping
BIN_SRC="${1:-bin/$APP}"          # optional arg: path to built binary
BIN_DST="/usr/local/bin/$APP"
CONFDIR="/etc/$APP"
ENVFILE="$CONFDIR/$APP.env"
HOSTS_FILE="$CONFDIR/hosts.txt"
UNIT_DST="/etc/systemd/system/$APP.service"
UNIT_SRC="$(dirname "$0")/../service/$APP.service"

# Defaults (can be edited later in $ENVFILE)
PORT="${PORT:-8090}"
INTERVAL="${INTERVAL:-2s}"
WINDOW="${WINDOW:-120}"

echo "==> Installing $APP"

# 1) service user
if ! id -u atping >/dev/null 2>&1; then
  sudo useradd --system --no-create-home --shell /usr/sbin/nologin atping
fi

# 2) binary
sudo install -Dm755 "$BIN_SRC" "$BIN_DST"

# 3) allow ICMP without root
if ! command -v setcap >/dev/null 2>&1; then
  sudo apt-get update -y
  sudo apt-get install -y libcap2-bin
fi
sudo setcap cap_net_raw+ep "$BIN_DST" || true

# 4) config dir + env + hosts
sudo install -d "$CONFDIR"

if [ ! -f "$ENVFILE" ]; then
  sudo tee "$ENVFILE" >/dev/null <<EOF
# at-ping runtime configuration
PORT=$PORT
INTERVAL=$INTERVAL
WINDOW=$WINDOW
HOSTS_FILE=$HOSTS_FILE
EOF
fi
sudo chown -R atping:atping "$CONFDIR"
sudo chmod 0644 "$ENVFILE"

if [ ! -f "$HOSTS_FILE" ]; then
  sudo tee "$HOSTS_FILE" >/dev/null <<'EOF'
# One host per line (blank lines and # comments ignored)
1.1.1.1
192.0.2.1
example.com
EOF
  sudo chown atping:atping "$HOSTS_FILE"
  sudo chmod 0644 "$HOSTS_FILE"
fi

# 5) systemd unit
sudo install -Dm644 "$UNIT_SRC" "$UNIT_DST"
sudo systemctl daemon-reload
sudo systemctl enable --now "$APP.service"

echo "==> $APP installed. Logs:"
echo "    sudo journalctl -u $APP -f"
