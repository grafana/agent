#!/usr/bin/env bash

EXTRA_MOUNTS=""

if [ -f /etc/machine-id ]; then
  EXTRA_MOUNTS="$EXTRA_MOUNTS -v /etc/machine-id:/etc/machine-id"
fi

if [ -d /dev/mapper ]; then
  EXTRA_MOUNTS="$EXTRA_MOUNTS -v /dev/mapper:/dev/mapper"
fi

k3d cluster create agent-k3d \
  --port 30080:80@loadbalancer \
  --api-port 50443 \
  -v /var/lib/k3d/agent-k3d/storage/:/var/lib/rancher/k3s/storage/ \
  $EXTRA_MOUNTS \
  --kubeconfig-update-default=true \
  --kubeconfig-switch-context=true \
  --wait
