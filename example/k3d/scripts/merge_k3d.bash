#!/usr/bin/env bash
#
# merge_k3d_config.bash
#
# Depends on https://github.com/kislyuk/yq:
#   pip install yq
set -e

command -v yq >/dev/null 2>&1 || {
  echo "yq is required but not found. install yq with 'pip install yq'."
  exit 1
}

NEW_KUBECONFIG=$(k3d get-kubeconfig --name='agent-k3d')

yq -s -y \ "(\
  .[0].clusters[0].name = \"agent-k3d\" | \
  .[0].users[0].name = \"agent-k3d\" | \
  .[0].contexts[0].context.cluster = \"agent-k3d\" | \
  .[0].contexts[0].context.user = \"agent-k3d\" | \
  .[0].contexts[0].name = \"agent-k3d\" | \
  .[0][\"current-context\"] = \"agent-k3d\" \
  ) | .[0]" $NEW_KUBECONFIG > ${NEW_KUBECONFIG}.bk

mv ${NEW_KUBECONFIG}.bk $NEW_KUBECONFIG

KUBECONFIG=${KUBECONFIG:=$HOME/.kube/config}
cp $KUBECONFIG ${KUBECONFIG}.bk
KUBECONFIG=$NEW_KUBECONFIG:${KUBECONFIG}.bk kubectl config view --raw > $KUBECONFIG
rm ${KUBECONFIG}.bk
