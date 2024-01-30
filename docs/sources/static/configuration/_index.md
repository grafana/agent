---
aliases:
- ../configuration/
- /docs/grafana-cloud/monitor-infrastructure/agent/static/configuration/
- /docs/grafana-cloud/send-data/agent/static/configuration/
canonical: https://grafana.com/docs/agent/latest/static/configuration/
description: Learn how to configure Grafana Agent in static mode
title: Configure static mode
weight: 300
---

# Configure static mode

The configuration of static mode is split across two places:

* A YAML file
* [Command-line flags][flags]

The YAML file is used to configure settings which are dynamic and can be
changed at runtime. The command-line flags then configure things which cannot
change at runtime, such as the listen port for the HTTP server.

This file describes the YAML configuration, which is usually in a file named `config.yaml`.

- [server_config][server]
- [metrics_config][metrics]
- [logs_config][logs]
- [traces_config][traces]
- [integrations_config][integrations]

The configuration of Grafana Agent is "stable," but subject to breaking changes
as individual features change. Breaking changes to configuration will be
well-documented.

## Updating configuration

The configuration file can be reloaded at runtime using the `/-/reload` API
endpoint or sending a SIGHUP signal to the process.

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

The configuration file can be reloaded at runtime. Read the [API documentation][api] for more information.

This functionality is in beta, and may have issues. Please open GitHub issues
for any problems you encounter.

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

Support contents and default values of `config.yaml`:

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

## Remote Configuration (Experimental)

An experimental feature for fetching remote configuration files over HTTP/S can be
enabled by passing the `-enable-features=remote-configs` flag at the command line.
With this feature enabled, you may pass an HTTP/S URL to the `-config.file` flag.

The following flags will configure basic auth for requests made to HTTP/S remote config URLs:
- `-config.url.basic-auth-user <user>`: the basic auth username
- `-config.url.basic-auth-password-file <file>`: path to a file containing the basic auth password

{{< admonition type="note" >}}
This beta feature is subject to change in future releases.
{{< /admonition >}}

{{% docs/reference %}}
[flags]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/flags"
[flags]: "/docs/grafana-cloud/ -> ./flags"
[server]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/server-config"
[server]: "/docs/grafana-cloud/ -> ./server-config"
[metrics]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/metrics-config"
[metrics]: "/docs/grafana-cloud/ -> ./metrics-config"
[logs]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/logs-config"
[logs]: "/docs/grafana-cloud/ -> ./logs-config"
[traces]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/traces-config"
[traces]: "/docs/grafana-cloud/ -> ./traces-config"
[integrations]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/configuration/integrations"
[integrations]: "/docs/grafana-cloud/ -> ./integrations"
[api]: "/docs/agent/ -> /docs/agent/<AGENT_VERSION>/static/api#reload-configuration-file-beta"
[api]: "/docs/grafana-cloud/ -> ../api#reload-configuration-file-beta"
{{% /docs/reference %}}
