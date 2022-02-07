+++
title = "snmp_exporter_config"
+++

# snmp_exporter_config

The `snmp_exporter_config` block configures the `snmp_exporter` integration,
which is an embedded version of
[`snmp_exporter`](https://github.com/prometheus/snmp-exporter). This allows for the collection of metrics from the SNMP devices.

To use snmp_exporter integration, enable it with autoscraping disabled. Then, configure scrape_configs to scrape metrics from SNMP targets:

```yaml

metrics:
  wal_directory: /tmp/wal
  configs:
    - name: snmp_targets
      scrape_configs:
        - job_name: 'snmp'
          static_configs:
            - targets: [192.168.15.1]  # First SNMP device.
              labels:
                __param_walk_params: public
            - targets: [192.168.15.2]  # Second SNMP device.
              labels:
                __param_walk_params: private
          metrics_path: /integrations/snmp_exporter/metrics
          params:
            module: [if_mib] # snmp module to use
          relabel_configs:
            - source_labels: [__address__]
              target_label: __param_target
            - source_labels: [__param_target]
              target_label: instance
            - replacement: 127.0.0.1:9090  # The SNMP exporter's real hostname:port.
              target_label: __address__ 

integrations:
  snmp_exporter:
    enabled: true
    scrape_integration: false
    walk_params:
      private:
        version: 2
        auth:
          community: secretpassword
      public:
        version: 2
        auth:
          community: public
server:
    http_listen_port: 9090
```


Full reference of options:

```yaml
  # Enables the snmp_exporter integration, allowing the Agent to automatically
  # collect metrics for the specified github objects.
  [enabled: <boolean> | default = false]

  # Sets an explicit value for the instance label when the integration is
  # self-scraped. Overrides inferred values.
  #
  # The default value for this integration is inferred from the hostname portion
  # of api_url.
  [instance: <string>]

  # Automatically collect metrics from this integration. If disabled,
  # the snmp_exporter integration will be run but not scraped and thus not
  # remote-written. Metrics for the integration will be exposed at
  # /integrations/snmp_exporter/metrics and can be scraped by an external
  # process.
  [scrape_integration: <boolean> | default = <integrations_config.scrape_integrations>]

  # How often should the metrics be collected? Defaults to
  # prometheus.global.scrape_interval.
  [scrape_interval: <duration> | default = <global_config.scrape_interval>]

  # The timeout before considering the scrape a failure. Defaults to
  # prometheus.global.scrape_timeout.
  [scrape_timeout: <duration> | default = <global_config.scrape_timeout>]

  # Allows for relabeling labels on the target.
  relabel_configs:
    [- <relabel_config> ... ]

  # Relabel metrics coming from the integration, allowing to drop series
  # from the integration that you don't care about.
  metric_relabel_configs:
    [ - <relabel_config> ... ]

  # How frequent to truncate the WAL for this integration.
  [wal_truncate_frequency: <duration> | default = "60m"]

  #
  # Exporter-specific configuration options
  #

  # SNMP configuration file with custom modules.
  # See https://github.com/prometheus/snmp_exporter#generating-configuration for more details how to generate custom snmp.yml file. 
  # If not defined, embeede snmp_exporter default set of modules is used.
  [config_file: <string> | default = ""]

  # Map of SNMP connection profiles that can be used to override default SNMP settings.
  walk_params:
    [ <string>: <walk_param> ... ]


```

## walk_param config

```yaml
  # SNMP version to use. Defaults to 2.
  # 1 will use GETNEXT, 2 and 3 use GETBULK.
  [version: <int> | default = 2]

  # How many objects to request with GET/GETBULK, defaults to 25.
  # May need to be reduced for buggy devices.
  [max_repetitions: <int> | default = 25]

  # How many times to retry a failed request, defaults to 3.
  [retries: <int> | default = 25]

  # Timeout for each individual SNMP request, defaults to 5s.
  [timeout: <duration> | default = 5s]

  auth:
    # Community string is used with SNMP v1 and v2. Defaults to "public".
    [community: <string> | default = "public"]

    # v3 has different and more complex settings.
    # Which are required depends on the security_level.
    # The equivalent options on NetSNMP commands like snmpbulkwalk
    # and snmpget are also listed. See snmpcmd(1).
    
    # Required if v3 is ued, no default. -u option to NetSNMP.
    [username: <string> | default = "user"] 

    # Defaults to noAuthNoPriv. -l option to NetSNMP.
    # Can be noAuthNoPriv, authNoPriv or authPriv.
    [security_level: <string> | default = "noAuthNoPriv"]

    # Has no default. Also known as authKey, -A option to NetSNMP.
    # Required if security_level is authNoPriv or authPriv.
    [password: <string> | default = ""]

    # MD5, SHA, SHA224, SHA256, SHA384, or SHA512. Defaults to MD5. -a option to NetSNMP.
    # Used if security_level is authNoPriv or authPriv.
    [auth_protocol: <string> | default = "MD5"]

    # DES, AES, AES192, or AES256. Defaults to DES. -x option to NetSNMP.
    # Used if security_level is authPriv.
    [priv_protocol: <string> | default = "DES"]
    
    # Has no default. Also known as privKey, -X option to NetSNMP.
    # Required if security_level is authPriv.
    [priv_password: <string> | default = ""]

    # Has no default. -n option to NetSNMP.
    # Required if context is configured on the device.  
    [context_name: <string> | default = ""]

```
