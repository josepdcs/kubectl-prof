#!/bin/bash
# Wrapper script for libbpf-tools profile to handle symbol resolution in containers
# When profiling a containerized process, we need to run the profiler in the target's
# mount namespace so it can access the binaries for symbol resolution.

set -e

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

# Copy the profile binary to the target's /tmp directory
PROC_ROOT="/proc/$PID/root"
TARGET_PROFILE="$PROC_ROOT/tmp/kubectl-prof-profile-$$"

if [ ! -d "$PROC_ROOT" ]; then
    echo "Error: Cannot access $PROC_ROOT" >&2
    exit 1
fi

# Copy profile binary to target's filesystem
cp /app/libbpf-profiler/profile "$TARGET_PROFILE"
chmod +x "$TARGET_PROFILE"

# Cleanup function
cleanup() {
    rm -f "$TARGET_PROFILE" 2>/dev/null || true
}
trap cleanup EXIT

# Use nsenter to run the profiler in the target's mount namespace
# The profile binary is now accessible at /tmp/kubectl-prof-profile-$$ in that namespace
exec nsenter -t "$PID" -m "/tmp/kubectl-prof-profile-$$" "${PROFILE_ARGS[@]}"
