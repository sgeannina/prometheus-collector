#!/bin/bash
#
# Builds and pushes a docker image into the specified ACR.
# Used as an EV2 script with assumptions about EV2 variables

#######################################
# Script is called as a function by EV2.
# Arguments:
#   - acr_name: name of ACR to push docker image to
#   - acr_subscription_id: sub ID of the ACR to push the docker image to
#   - image_name: name of the docker image
#   - docker_tag: tag of the docker image
#   - ame_tenant_id: tenant where the ACR lives in
#   - RolloutId: EV2 RolloutId. https://ev2docs.azure.net/features/extensibility/shell/artifacts.html#rollout-parameters
# Returns:
#   error code
#######################################

set -eu

echo "Installing jq"
apt-get update
apt-get install jq -y
echo "Finished installing jq"

apt-get install gettext-base -y

echo "Setting gcr_endpoint: ${gcr_endpoint}..."
GCR_ENDPOINT="${gcr_endpoint}" envsubst < Dockerfile > temp.txt
mv temp.txt Dockerfile

echo "Logging into Azure..."
az login --identity

echo "Logged in, pushing image"
az acr build --subscription $acr_subscription_id --registry $acr_name --image public/aks/hcp/${image_name}:${docker_tag} .

echo "Successfully pushed new image to ACR"
