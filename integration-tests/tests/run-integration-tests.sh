#!/bin/bash
set -e

need_cleanup=1
failed=0
logfile="grafana-agent-flow.log"

cleanup() {
    if [ "$need_cleanup" -eq "1" ]; then
        echo "Cleaning up..."
        if [ "$failed" -eq "1" ]; then
            echo "Capturing grafana-agent-flow logs due to a failure..."
            cat "$logfile"
            docker-compose down
        fi
        kill $AGENT_PID || true
        rm -rf data-agent
        rm -f "$logfile"
    fi
}

success() {
    need_cleanup=0
    echo "All integration tests passed!"
    exit 0
}

make -C ../.. agent-flow
AGENT_BINARY_PATH="../../../build/grafana-agent-flow"

trap cleanup EXIT ERR

docker-compose up -d

while read -r test_dir; do
    pushd "$test_dir"
    "$AGENT_BINARY_PATH" run config.river > "$logfile" 2>&1 &
    AGENT_PID=$!
    if ! go test; then
        failed=1
        exit 1
    fi
    cleanup
    popd
done < <(find . -maxdepth 1 -type d ! -path .)

docker-compose down

success
