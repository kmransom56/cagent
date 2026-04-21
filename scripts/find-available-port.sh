#!/usr/bin/env bash
# Find Available Port in 11000-12000 Range
# Usage: ./scripts/find-available-port.sh [startPort] [endPort]

set -euo pipefail

START_PORT="${1:-11000}"
END_PORT="${2:-12000}"

echo "Finding available port in range ${START_PORT}-${END_PORT}..." >&2

for port in $(seq "$START_PORT" "$END_PORT"); do
    if ! (echo >/dev/tcp/localhost/"$port") 2>/dev/null; then
        echo "✓ Found available port: ${port}" >&2
        echo "$port"
        exit 0
    fi
done

echo "✗ No available ports in range ${START_PORT}-${END_PORT}" >&2
exit 1
