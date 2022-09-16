#!/usr/bin/env bash

ROOT=$(git rev-parse --show-toplevel)


# Generate objects and controllers for our CRDs
cd $ROOT/pkg/operator/apis/monitoring/v1alpha1
controller-gen object paths=.
controller-gen crd:crdVersions=v1 paths=. output:crd:dir=$ROOT/production/operator/crds

# Generate CRDs for prometheus-operator.
PROM_OP_DEP_NAME="github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
PROM_OP_DIR=$(go list -f '{{.Dir}}' $PROM_OP_DEP_NAME)

cd $PROM_OP_DIR
controller-gen crd:crdVersions=v1 paths=.  output:crd:dir=$ROOT/production/operator/crds

# Generate API docs for agent-operator
gen-crd-api-reference-docs -config $ROOT/tools/gen-crd-docs/config.json -api-dir "github.com/grafana/agent/pkg/operator/apis/monitoring/" -out-file $ROOT/docs/sources/operator/api.md -template-dir $ROOT/tools/gen-crd-docs/template

# Remove known Prometheus-Operator CRDS we don't generate. (An allowlist would
# be better here, but rfratto's bash skills are bad.)
rm -f $ROOT/production/operator/crds/monitoring.coreos.com_alertmanagers.yaml
rm -f $ROOT/production/operator/crds/monitoring.coreos.com_prometheuses.yaml
rm -f $ROOT/production/operator/crds/monitoring.coreos.com_prometheusrules.yaml
rm -f $ROOT/production/operator/crds/monitoring.coreos.com_thanosrulers.yaml
