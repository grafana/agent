#!/usr/bin/env bash

CHART_PATH="operations/helm/charts/grafana-agent"
OUTPUT_PATH="operations/helm/tests"

rm -rf $OUTPUT_PATH

CHART_NAME=$(basename $CHART_PATH)
TESTS=$(find "${CHART_PATH}/tests" -name "*.values.yaml")

for FILEPATH in $TESTS; do
  FILENAME=$(basename $FILEPATH)
  TESTNAME=${FILENAME%.values.yaml}

  helm template --namespace default --debug ${CHART_NAME} ${CHART_PATH} -f ${FILEPATH} --output-dir ${OUTPUT_PATH}/${TESTNAME} --set '$chart_tests=true'
done
