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
        fi
        kill $AGENT_PID || true
        pwd
        docker-compose down
        rm -rf data-agent
        rm -f "$logfile"
    fi
}

success() {
    need_cleanup=0
    echo "All integration tests passed!"
    exit 0
}

trap cleanup EXIT ERR

make -C ../.. agent-flow

for i in {1..10}; do
    while read -r test_dir; do
        pushd "$test_dir"
        docker-compose up -d
        sleep 5
        ../../../build/grafana-agent-flow run config.river > "$logfile" 2>&1 &
        AGENT_PID=$!
        if ! go test; then
            failed=1
            exit 1
        fi
        cleanup
        popd
    done < <(find . -maxdepth 1 -type d ! -path .)
done

success
