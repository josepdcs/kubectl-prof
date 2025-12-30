#!/bin/sh
# Wrapper script for perf to ensure it can be found
# This is needed because with hostPID, the PATH may not include /app

if [ -x "/usr/bin/perf" ]; then
    exec /usr/bin/perf "$@"
elif [ -x "/app/perf" ] && [ "$(realpath /app/perf)" != "$(realpath $0)" ]; then
    exec /app/perf "$@"
else
    echo "Error: perf not found in /usr/bin/perf" >&2
    exit 1
fi

