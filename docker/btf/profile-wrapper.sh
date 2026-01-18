#!/bin/bash
# Wrapper script for libbpf-tools profile to handle symbol resolution in containers
# When profiling a containerized process, we need to run the profiler in the target's
# mount namespace so it can access the binaries for symbol resolution.

PID=""
PROFILE_ARGS=()

# Parse arguments to extract the PID
while [[ $# -gt 0 ]]; do
    case "$1" in
        -p)
            PID="$2"
            PROFILE_ARGS+=("$1" "$2")
            shift 2
            ;;
        *)
            PROFILE_ARGS+=("$1")
            shift
            ;;
    esac
done

if [ -z "$PID" ]; then
    echo "Error: PID not specified" >&2
    exit 1
fi

# Run the profiler directly without nsenter
# With hostPID: true, the profiler can access /proc/<pid>/maps and /proc/<pid>/root/*
# The kernel automatically provides symbol resolution by reading the target's binaries
# via /proc/<pid>/root/<path-to-binary>
exec /app/libbpf-profiler/profile "${PROFILE_ARGS[@]}"
