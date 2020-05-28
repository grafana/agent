#!/usr/bin/env bash

k3d create \
  --name agent-k3d \
  --publish 30080:30080 \
  --api-port 50443 \
  -v /var/lib/k3d/agent-k3d/storage/:/var/lib/rancher/k3s/storage/
