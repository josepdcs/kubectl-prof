#!/usr/bin/env bash
set -eou pipefail

THIS_SCRIPT="test/minikube-lab/build_and_push_stupid_apps.sh"
APPS_DIR="test/stupid-apps"

. $(dirname "$0")/init.sh

process_images_dir "${APPS_DIR}"

show_current_images_in_minikube
