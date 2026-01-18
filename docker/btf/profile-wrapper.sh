#!/bin/bash
set -e  # Exit on error

# Wrapper script for libbpf-tools profile
# Simply passes all arguments directly to the profile binary
exec /app/libbpf-profiler/profile "$@"
