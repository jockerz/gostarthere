#!/bin/sh
#
set -eu

# Create media directories
mkdir -p /app/media/avatar

# Start worker in background
/app/main worker &
WORKER_PID=$!

# Start API server in background
/app/main api &
API_PID=$!

terminate() {
    kill -TERM "$API_PID" "$WORKER_PID" 2>/dev/null || true
    wait "$API_PID" "$WORKER_PID" 2>/dev/null || true
}

trap terminate TERM INT

wait "$API_PID"
STATUS=$?

kill -TERM "$WORKER_PID" 2>/dev/null || true
wait "$WORKER_PID" 2>/dev/null || true

exit "$STATUS"
