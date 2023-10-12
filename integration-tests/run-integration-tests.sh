#!/bin/bash
set -e

fail_flag_file="/tmp/test_fail_flag"
rm -f "$fail_flag_file"
tmp_dir="temp_logs"

cleanup() {
    if [ -d "$tmp_dir" ]; then
        for tmp_log in "$tmp_dir"/*; do
            echo "$tmp_log"
            if grep -q "FAIL" "$tmp_log"; then
                echo "Failure detected in $tmp_log:"
                cat "$tmp_log"
            fi
            rm -f "$tmp_log"
        done
        rmdir "$tmp_dir"
    fi
    docker-compose down
}

success() {
    echo "All integration tests passed!"
    exit 0
}

#make -C .. agent-flow
AGENT_BINARY_PATH="../../../build/grafana-agent-flow"

docker-compose up -d

mkdir -p "$tmp_dir"

counter=0

# Run tests in parallel
while read -r test_dir; do
    (
        pushd "$test_dir"
        dir_name=$(basename "$test_dir")
        agent_logfile="../../${tmp_dir}/${dir_name}_agent.log"
        test_logfile="../../${tmp_dir}/${dir_name}_test.log"
        "$AGENT_BINARY_PATH" run config.river > "$agent_logfile" 2>&1 &
        AGENT_PID=$!
        if ! go test >> "$test_logfile" 2>&1; then
            echo "FAIL" >> "$test_logfile"
            touch "$fail_flag_file"
        fi
        # Concatenate the log files into one if desired.
        cat "$agent_logfile" >> "$test_logfile"
        rm "$agent_logfile"

        rm -rf data-agent
        kill $AGENT_PID || true
        popd
    ) &

    # Increment the counter
    counter=$((counter+1))

    # If 5 tests are running, wait for them to finish
    if [ "$counter" -eq 5 ]; then
        wait
        counter=0
    fi
done < <(find ./tests -maxdepth 1 -type d ! -path ./tests)

wait

pwd

cleanup

if [ -f "$fail_flag_file" ]; then
    rm "$fail_flag_file"
    exit 1
else
    success
fi
