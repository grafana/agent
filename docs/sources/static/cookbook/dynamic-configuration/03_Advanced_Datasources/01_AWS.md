---
aliases:
- ../../../dynamic-configuration/aws/
title: Querying AWS
weight: 300
---

# 01 AWS

The AWS datasource assumes that you have appropriate credentials and environment variables set to access AWS resources. The custom fork of gomplate adds a new command to the existing AWS commands.

Unfortunately there is not a specific docker command but generic examples are below.

## Looping

[agent-1.yml](01_assets/agent-1.yml)

```yaml
server:
  log_level: debug
metrics:
  wal_directory: /tmp/grafana-agent-normal
  global:
    scrape_interval: 60s
    remote_write:
      - url: https://prometheus-us-central1.grafana.net/api/prom/push
        basic_auth:
          username: xyz
          password: secretpassword
  configs:
    - name: default
      scrape_configs:
      {{ range $index , $value := aws.EC2Query "tag:service=webhost" -}}
      - job_name: {{ $value.InstanceId }}
        static_configs:
          - targets:
              - {{ $value.PrivateDnsName }}
        {{ end -}}
```

The `aws.EC2Query` command is a new command added for Grafana Agent and takes a string in the [DescribeInstances](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstances.html) format

## Final

[final.yml](01_assets/final.yml)
