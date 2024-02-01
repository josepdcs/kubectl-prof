#!/usr/bin/env bash
set -eou pipefail

declare -A minikube_profiles
minikube_profiles=(
  ["1"]="minikube-virtualbox-crio"
  ["2"]="minikube-virtualbox-containerd"
  ["3"]="minikube-docker-crio"
  ["4"]="minikube-docker-containerd"
  ["5"]="minikube-qemu2-containerd"
  ["6"]="minikube-qemu2-crio"
)

function read_and_set_minikube_profile() {
  print_minikube_profiles

  read -p "Enter one [1/2/3/4/5]: " input
  case "$input" in
  1|2|3|4|5|6)
    MINIKUBE_PROFILE=${minikube_profiles[$input]}
    ;;
  *)
    echo "Invalid response"
    read_and_set_minikube_profile
    ;;
  esac

  echo "💡 For skipping this step, set env var to MINIKUBE_PROFILE=$MINIKUBE_PROFILE"
  echo ""
}

function print_minikube_profiles() {
  echo ""
  echo "Available minikube profiles:"
  echo -e "ID\tMINIKUBE PROFILE"
  for minikube_profiles in ${!minikube_profiles[@]}; do
    echo -e "$minikube_profiles\t${minikube_profiles[$minikube_profiles]}"
  done |
    sort -n
  echo ""
}

function check_for_minikube_profile() {
  IFS="-" read -a elements <<< $MINIKUBE_PROFILE

  driver=${elements[1]}
  runtime=${elements[2]}

  if [[ ! "${minikube_profiles[*]}" =~ "$MINIKUBE_PROFILE" ]]; then
    echo "😞 Unsupported minikube profile: $MINIKUBE_PROFILE"
    exit
  fi
}

# shellcheck disable=SC2046,SC2086
function build_docker_image_and_push_to_minikube() {
  # Receives the directory where we will build the image
  local workDir=${1}
  local project=${2}
  local image=${3}
  local fullImageName=${REGISTRY}/${project}/${image}:latest
  local dockerfile="${workDir}/Dockerfile"

  #Building image: localhost/stupid-apps/clang:latest. Dockerfile: test/stupid-apps/clang/Dockerfile
  echo "Building image: ${fullImageName}. Dockerfile: ${dockerfile}"

  docker build -t ${fullImageName} -f "${dockerfile}" .

  echo "Loading image ${fullImageName} to $MINIKUBE_PROFILE"
  minikube image load ${fullImageName} -p $MINIKUBE_PROFILE --overwrite=true --daemon

  echo "Removing image ${fullImageName} from local registry"
  docker rmi $(docker images --filter "dangling=true" -q --no-trunc) ${fullImageName} || true
}

# shellcheck disable=SC2046,SC2086
function build_docker_image_and_push_to_docker() {
  # Receives the directory where we will build the image
  local workDir=${1}
  local user=${2}
  local image=${3}
  local fullImageName=docker.io/${user}/${image}:latest
  local dockerfile="${workDir}/Dockerfile"

  echo "Building image: ${fullImageName}. Dockerfile: ${dockerfile}"

  docker build -t ${fullImageName} -f "${dockerfile}" .

  echo "Pushing image ${fullImageName} to Docker"
  docker push ${fullImageName}

  echo "Removing image ${fullImageName} from local registry"
  docker rmi $(docker images --filter "dangling=true" -q --no-trunc) ${fullImageName} || true
}

function process_images_dir() {
  local file=${1}

  # Build docker images, but only for directories (this is to skip other files such a README.md etc...)
  if [[ -d "${file}" ]]; then
    # The first directory is the "name" of the image. E:g:
    #├── init (name of the image)
    #   └── Dockerfile

    for imageName in "${file}"/*; do
      if [[ -d "${imageName}" ]]; then
        image=$(basename "${imageName}")
        project=$(basename "${file}")
        build_docker_image_and_push_to_minikube "${imageName}" "${project}" "${image}"
      fi
    done
  fi
}

function show_current_images_in_minikube() {
  echo "====================================================="
  echo "Current images in $MINIKUBE_PROFILE cluster:"
  echo "====================================================="
  minikube image ls -p $MINIKUBE_PROFILE
}
