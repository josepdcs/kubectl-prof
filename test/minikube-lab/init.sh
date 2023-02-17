#!/usr/bin/env bash
set -eou pipefail

. $(dirname "$0")/functions.sh

REGISTRY=${REGISTRY:-"localhost"}
MINIKUBE_PROFILE=${MINIKUBE_PROFILE:-""}

if ! [ -x "$(command -v minikube)" ]; then
  echo "Error: «minikube» is not installed. Go to: https://minikube.sigs.k8s.io/docs/start/"
  exit 1
fi

if [[ "$PWD" != *kubectl-prof ]]; then
  echo "Wrong location: go to root folder 'kubectl-prof' and run '$THIS_SCRIPT' from there."
  exit 1
fi

if [ "$MINIKUBE_PROFILE" == "" ]; then
  echo "No minikube profile defined in MINIKUBE_PROFILE env var"
  read_and_set_minikube_profile
fi

check_for_minikube_profile

minikube profile $MINIKUBE_PROFILE


