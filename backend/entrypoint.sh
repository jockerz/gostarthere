#!/bin/sh
set -e

# Create media directories
mkdir -p /app/media/avatar

# Start worker in background
/app/main worker &

# Start API server (foreground)
exec /app/main api
