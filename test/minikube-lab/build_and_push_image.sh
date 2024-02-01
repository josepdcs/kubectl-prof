#!/usr/bin/env bash
set -eou pipefail

THIS_SCRIPT="test/minikube-lab/build_and_push_agent.sh"

. $(dirname "$0")/init.sh

if [ "$REGISTRY" == "docker.io" ]; then
  build_docker_image_and_push_to_docker "${1}" "${2}" "${3}"
else
  build_docker_image_and_push_to_minikube "${1}" "${2}" "${3}"
fi

show_current_images_in_minikube
