You can start agent flow by passing in 
```
--agent.flow
--config.file=/Users/mdurham/Utils/agent_flow_configs/agent_flow_prom.yml
```

`pkg/agentflow/actorsystem/system.go` is the root for the actor system. Config is added to the config and hopefully you can follow the existing components.


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
      username: 12345
      password: password
      url: https://prometheus-us-central1.grafana.net/api/prom/push
  - name: agent_logs
    agent_logs: {}
    outputs:
      - filewriter
  - name: filewriter
    log_file_writer:
      path: /tmp/log/agent_flow_configs/logs.txt
  - name: github
    github:
      enable_endpoint: true
      repositories:
        - grafana/agent
    outputs:
      - rw


```