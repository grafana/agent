#!/usr/bin/env bash

set +x

CHART_PATH="charts/grafana-agent"
OUTPUT_PATH="tests"

CHART_PATH_FROM_REPO_ROOT="operations/helm/${CHART_PATH}"
OUTPUT_PATH_FROM_REPO_ROOT="operations/helm/${OUTPUT_PATH}"

CURRENT_DIR_NAME=${PWD##*/}

if [ $CURRENT_DIR_NAME != "helm" ]
then
  CHART_PATH=$CHART_PATH_FROM_REPO_ROOT
  OUTPUT_PATH=$OUTPUT_PATH_FROM_REPO_ROOT
fi

rm -rf $OUTPUT_PATH

CHART_NAME=$(basename $CHART_PATH)
TESTS=$(find "${CHART_PATH}/tests" -name "*.values.yaml")

for FILEPATH in $TESTS; do
  FILENAME=$(basename $FILEPATH)
  TESTNAME=${FILENAME%.values.yaml}

  echo "Render with file ${FILEPATH}"
  helm template --namespace default --debug ${CHART_NAME} ${CHART_PATH} -f ${FILEPATH} --output-dir ${OUTPUT_PATH}/${TESTNAME}
done
