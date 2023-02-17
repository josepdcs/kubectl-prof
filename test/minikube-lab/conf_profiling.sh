#!/usr/bin/env bash
set -eou pipefail

. $(dirname "$0")/init.sh

kubectl apply -f test/minikube-lab/manifests/profiling.yml