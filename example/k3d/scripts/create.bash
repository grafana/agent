#!/usr/bin/env bash

k3d cluster create agent-k3d \
  --port 30080:80@loadbalancer \
  --api-port 50443 \
  -v /var/lib/k3d/agent-k3d/storage/:/var/lib/rancher/k3s/storage/ \
  -v /etc/machine-id:/etc/machine-id \
  --kubeconfig-update-default=true \
  --kubeconfig-switch-context=true \
  --wait
