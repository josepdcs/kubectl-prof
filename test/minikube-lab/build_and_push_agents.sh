#!/usr/bin/env bash
set -eou pipefail

THIS_SCRIPT="test/minikube-lab/build_and_push_agents.sh"
CONTPROF_IMAGE_DIR="docker"

. $(dirname "$0")/init.sh

process_images_dir "${CONTPROF_IMAGE_DIR}"

show_current_images_in_minikube
