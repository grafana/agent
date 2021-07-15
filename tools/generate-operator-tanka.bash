#!/usr/bin/env bash

ROOT=$(git rev-parse --show-toplevel)
STARTDIR=$(pwd)

rm -rf /tmp/k8s-tmp
git clone https://github.com/jsonnet-libs/k8s /tmp/k8s-tmp
cd /tmp/k8s-tmp

mkdir -p libs/grafana-agent-operator
rm -f libs/grafana-agent-operator/crd.yaml
cat $ROOT/production/operator/crds/*.yaml >> libs/grafana-agent-operator/crd.yaml

cat <<'EOF' >libs/grafana-agent-operator/config.jsonnet
local config = import 'jsonnet/config.jsonnet';

config.new(
  name='grafana-agent-operator',
  specs=[
    {
      output: 'unstable',
      openapi: 'http://localhost:8001/openapi/v2',
      prefix: '^com\\.(coreos|grafana)\\.monitoring\\..*',
      crds: ['file:///config/crd.yaml'],
      localName: 'grafana_agent_operator',
    },
  ],
) {
  'skel/README.md': |||
    # Grafana Agent Operator Jsonnet library

    This library assists with generating Grafana Agent Operator custom
    resources. It is generated with [`k8s`](https://github.com/jsonnet-libs/k8s).

    To update the library, run `make operator-tanka-library` from the root of the
    [Grafana Agent repository](https://github.com/grafana/agent).
  |||
}
EOF

make run INPUT_DIR=libs/grafana-agent-operator

cd $STARTDIR

OUTPUT_DIR=$ROOT/production/tanka/grafana-agent-operator
rm -rf $OUTPUT_DIR && mkdir $OUTPUT_DIR
cp -rv /tmp/k8s-tmp/gen/github.com/jsonnet-libs/grafana-agent-operator-lib/ $OUTPUT_DIR/
rm -rf /tmp/k8s-tmp
