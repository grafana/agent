+++
title = "Configure Grafana Agent"
weight = 300
+++

# Configure Grafana Agent

The Grafana Agent is configured in a YAML file (usually called
`agent.yaml`) which contains information on the Grafana Agent and its
Prometheus instances.

- [server_config]({{< relref "./server-config.md" >}})
- [prometheus_config]({{< relref "./prometheus-config.md" >}})
- [loki_config]({{< relref "./loki-config.md" >}})
- [tempo_config]({{< relref "./tempo-config.md" >}})
- [integrations_config]({{< relref "./integrations/_index.md" >}})

## Variable substitution

You can use environment variables in the configuration file to set values that
need to be configurable during deployment. To enable this functionality, you
must pass `-config.expand-env` as a command-line flag to the Agent.

To refer to an environment variable in the config file, use:

```
${VAR}
```

Where VAR is the name of the environment variable.

Each variable reference is replaced at startup by the value of the environment
variable. The replacement is case-sensitive and occurs before the YAML file is
parsed. References to undefined variables are replaced by empty strings unless
you specify a default value or custom error text.

To specify a default value, use:

```
${VAR:-default_value}
```

Where default_value is the value to use if the environment variable is
undefined. The full list of supported syntax can be found at Drone's
[envsubst repository](https://github.com/drone/envsubst).

## Reloading (beta)

The configuration file can be reloaded at runtime. Read the [API
documentation](../api.md#reload-configuration-file-beta) for more information.

This functionality is in beta, and may have issues. Please open GitHub issues
for any problems you encounter.

A reload-only HTTP server can be started to safely reload the system. To start
this, provide `--reload-addr` and `--reload-port` as command line flags.
`reload-port` must be set to a non-zero port to launch the reload server. The
reload server is currently HTTP-only and supports no other options; it does not
read any values from the `server` block in the config file.

While `/-/reload` is enabled on the primary HTTP server, it is not recommended
to use it, since changing the HTTP server configuration will cause it to
restart.

## File format

To specify which configuration file to load, pass the `-config.file` flag at
the command line. The file is written in the [YAML
format](https://en.wikipedia.org/wiki/YAML), defined by the scheme below.
Brackets indicate that a parameter is optional. For non-list parameters the
value is set to the specified default.

Generic placeholders are defined as follows:

- `<boolean>`: a boolean that can take the values `true` or `false`
- `<int>`: any integer matching the regular expression `[1-9]+[0-9]*`
- `<duration>`: a duration matching the regular expression `[0-9]+(ns|us|µs|ms|[smh])`
- `<labelname>`: a string matching the regular expression `[a-zA-Z_][a-zA-Z0-9_]*`
- `<labelvalue>`: a string of unicode characters
- `<filename>`: a valid path relative to current working directory or an
    absolute path.
- `<host>`: a valid string consisting of a hostname or IP followed by an optional port number
- `<string>`: a regular string
- `<secret>`: a regular string that is a secret, such as a password

Support contents and default values of `agent.yaml`:

```yaml
# Configures the server of the Agent used to enable self-scraping.
[server: <server_config>]

# Configures Prometheus instances.
[prometheus: <prometheus_config>]

# Configures Loki log collection.
[loki: <loki_config>]

# Configures Tempo trace collection.
[tempo: <tempo_config>]

# Configures integrations for the Agent.
[integrations: <integrations_config>]
```
