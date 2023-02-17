#!/usr/bin/env bash
set -eou pipefail

. $(dirname "$0")/functions.sh
. $(dirname "$0")/start_clusters.sh
. $(dirname "$0")/build_and_push_stupid_apps.sh
. $(dirname "$0")/build_and_push_agents.sh
. $(dirname "$0")/conf_profiling.sh
. $(dirname "$0")/deploy_stupid_apps.sh

echo "====================================================================="
echo "Current pods in minikube $MINIKUBE_PROFILE, namespace stupid-apps:"
echo "====================================================================="
kubectl get pods --namespace stupid-apps

