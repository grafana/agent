for f in agent/instance-configs/*.yaml; do
  BASENAME=$(basename $f)
  CONFIG_NAME=${BASENAME%.yaml}
  cat $f | curl -XPUT -H "Content-Type: text/yaml" --data-binary @- localhost:12345/agent/api/v1/config/${CONFIG_NAME}
done
