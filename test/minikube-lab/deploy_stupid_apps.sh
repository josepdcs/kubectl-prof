#!/usr/bin/env bash
set -eou pipefail

THIS_SCRIPT="test/minikube-lab/deploy_stupid_apps.sh"
APPS_DIR="test/stupid-apps*"

. $(dirname "$0")/init.sh

kubectl apply -f test/stupid-apps/stupid-apps.yml

for file in $APPS_DIR/*; do
  if [[ -d "${file}" ]]; then
    kubectl apply -f  ${file}/deployment.yml
  fi
done


