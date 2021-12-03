+++
title = "Configure Grafana Agent"
weight = 300
+++

# Configure Grafana Agent

The Grafana Agent is configured in a YAML file (usually called
`agent.yaml`) which contains information on the Grafana Agent and its
metrics instances.

- [server_config]({{< relref "./server-config" >}})
- [metrics_config]({{< relref "./metrics-config" >}})
- [logs_config]({{< relref "./logs-config.md" >}})
- [traces_config]({{< relref "./traces-config" >}})
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

### Regex capture group references

When using `-config.expand-env`, `VAR` must be an alphanumeric string with at
least one non-digit character. If `VAR` is a number, the expander will assume
you're trying to use a regex capture group reference, and will coerce the result
to be one.

This means references in your config file like `${1}` will remain
untouched, but edge cases like `${1:-default}` will also be coerced to `${1}`,
which may be slightly unexpected.

## Reloading (beta)

The configuration file can be reloaded at runtime. Read the [API
documentation]({{< relref "../api#reload-configuration-file-beta" >}}) for more information.

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

# Configures metric collection.
# In previous versions of the agent, this field was called "prometheus".
[metrics: <metrics_config>]

# Configures log collection.
# In previous versions of the agent, this field was called "loki".
[logs: <logs_config>]

# Configures Traces trace collection.
# In previous versions of the agent, this field was called "tempo".
[traces: <traces_config>]

# Configures integrations for the Agent.
[integrations: <integrations_config>]
```

## Remote Configuration (Beta)

An experimental feature for fetching remote configuration files over HTTP/S can be
enabled by passing the `-experiment.config-urls.enable` flag at the command line.
With this feature enabled, you may pass an HTTP/S URL to the `-config.file` flag.

The following flags will configure basic auth for requests made to HTTP/S remote config URLs:
- `-config.url.basic-auth-user <user>`: the basic auth username
- `-config.url.basic-auth-password-file <file>`: path to a file containing the basic auth password

Note that this beta feature is subject to change in future releases.
