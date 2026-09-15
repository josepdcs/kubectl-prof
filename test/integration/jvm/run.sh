#!/usr/bin/env bash
#
# Runs the JVM profiler integration suite against a non-root JVM. The agent container
# shares the host PID namespace so procfs resolves the target's mounts, while a bind
# mount provides the host-side filesystem path used to stage the profiler library.
#
# Requires Docker and network access to pull the configured images.

set -euo pipefail

AGENT_IMAGE=${AGENT_IMAGE:-josepdcs/kubectl-prof:2.2.0-jvm}
JDK_IMAGE=${JDK_IMAGE:-eclipse-temurin:21-jdk}
TARGET_UID=${TARGET_UID:-65532}
STAND=${STAND:-/tmp/kubectl-prof-integration}

here=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$here/../../.." && pwd)
arch=$(docker version --format '{{.Server.Arch}}')
binary=$repo/jvm.test
container=""

remove_stand() {
  docker run --rm -v /tmp:/host-tmp alpine:3.20 \
    rm -rf "/host-tmp/$(basename "$STAND")" >/dev/null 2>&1 || true
}

cleanup() {
  [[ -n $container ]] && docker rm -f "$container" >/dev/null 2>&1 || true
  container=""
}

trap 'cleanup; remove_stand; rm -f "$binary"' EXIT

echo "building the test binary for linux/$arch"
(cd "$repo" && GOOS=linux GOARCH="$arch" go test -c -tags integration -o "$binary" ./internal/agent/profiler/jvm/)

run_suite() {
  local label=$1
  shift
  local target_args=("$@")

  echo
  echo "################ $label ################"
  remove_stand

  # Expanded this way so an empty array does not trip set -u under bash 3.2.
  container=$(docker run -d --rm ${target_args[@]+"${target_args[@]}"} \
    -u "$TARGET_UID:$TARGET_UID" \
    -v "$STAND/kubectl-prof:/kubectl-prof" \
    -v "$here/Spin.java:/work/Spin.java:ro" \
    "$JDK_IMAGE" java /work/Spin.java)

  # Some JVMs create the attach socket lazily, so this wait is best-effort.
  local ready=""
  for _ in $(seq 40); do
    if docker exec "$container" sh -c 'ls /tmp/.java_pid* >/dev/null 2>&1'; then
      ready=yes
      break
    fi
    sleep 1
  done
  if [[ -z $ready ]]; then
    sleep 5
  fi

  local pid
  pid=$(docker inspect -f '{{.State.Pid}}' "$container")
  echo "target pid on the host: $pid (uid $TARGET_UID)"

  docker run --rm --pid=host --privileged \
    -e TARGET_PID="$pid" \
    -e TARGET_UID="$TARGET_UID" \
    -e TARGET_ROOTFS=/target-root \
    -v "$STAND:/target-root" \
    -v "$binary:/jvm.test:ro" \
    "$AGENT_IMAGE" /jvm.test -test.v -test.timeout=10m -test.run '^TestIntegration_'

  cleanup
}

run_suite "non-root JVM, default root filesystem"

run_suite "non-root JVM, read-only root filesystem with a tmpfs /tmp" \
  --read-only --tmpfs "/tmp:rw,exec,size=128m"

echo
echo "integration suite passed"
