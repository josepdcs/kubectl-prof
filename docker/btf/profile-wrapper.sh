#!/bin/bash
# Wrapper script for libbpf-tools profile to handle symbol resolution in containers
# The profiler needs to access target process binaries at their original paths

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

# Read the process's memory maps to find all loaded binaries
# and create symbolic links to make them accessible at their original paths
while IFS= read -r line; do
    # Extract the path from the maps file (last column)
    path=$(echo "$line" | awk '{print $NF}')
    
    # Skip if not an absolute path or if it's a special file
    if [[ ! "$path" =~ ^/ ]] || [[ "$path" =~ ^\[.*\]$ ]]; then
        continue
    fi
    
    # Check if the file exists in the target's root
    if [ -e "/proc/${PID}/root${path}" ]; then
        # Create directory structure if needed
        dir=$(dirname "$path")
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir" 2>/dev/null || true
        fi
        
        # Create symbolic link from the expected path to the actual location in /proc
        if [ ! -e "$path" ]; then
            ln -sf "/proc/${PID}/root${path}" "$path" 2>/dev/null || true
        fi
    fi
done < "/proc/${PID}/maps"

# Run the profiler - it should now find binaries at their original paths
exec /app/libbpf-profiler/profile "${PROFILE_ARGS[@]}"
