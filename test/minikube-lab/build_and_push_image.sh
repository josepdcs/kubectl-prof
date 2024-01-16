#!/usr/bin/env bash
set -eou pipefail

THIS_SCRIPT="test/minikube-lab/build_and_push_agent.sh"

. $(dirname "$0")/init.sh

build_docker_image_and_push_to_minikube "${1}" "${2}" "${3}"

show_current_images_in_minikube
