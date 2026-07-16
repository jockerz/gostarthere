#!/bin/sh
set -e

# Create media directories
mkdir -p /app/media/avatar

# Trap SIGTERM/SIGINT and forward to child processes
trap 'kill -TERM $API_PID $WORKER_PID 2>/dev/null; wait' TERM INT

# Start worker in background
/app/main worker &
WORKER_PID=$!

# Start API server in background
/app/main api &
API_PID=$!

# Wait for either process to exit
wait $API_PID $WORKER_PID
