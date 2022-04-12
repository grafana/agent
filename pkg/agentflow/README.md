Example config
```yaml
nodes:
- name: generator
  outputs: 
  - filter
  metric_generator:
    spawn_interval: 10s
- name: filter
  outputs:
  - filter2
  metric_filter:
    filters:
    - action: add_label
      add_label: test_label
      add_value: test
- name: filter2
  outputs:
  - rw
  metric_filter:
    filters:
    - action: add_label
      add_label: filter2_label
      add_value: "this is from filter 2"
- name: rw
  prometheus_remote_write:
    wal_dir: /tmp/agent-flow/wal
    username: <username>
    password: <password>
    url: https://prometheus-us-central1.grafana.net/api/prom/push
- name: agent_logs
  agent_logs: {}
  outputs:
  - filewriter
- name: filewriter
  log_file_writer:
    path: /path_to_logs/logs.txt
- name: github
  github:
    repositories:
    - grafana/agent
  outputs:
  - rw

```