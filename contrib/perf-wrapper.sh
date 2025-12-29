#!/bin/sh
# Wrapper script for perf to ensure it can be found
# This is needed because with hostPID, the PATH may not include /app

if [ -x "/app/perf" ]; then
    exec /app/perf "$@"
elif [ -x "/usr/bin/perf" ]; then
    exec /usr/bin/perf "$@"
else
    echo "Error: perf not found in /app/perf or /usr/bin/perf" >&2
    exit 1
fi

