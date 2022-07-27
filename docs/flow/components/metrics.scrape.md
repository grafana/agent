# metrics.scrape
The `metrics.scrape` component configures a metrics scraping job for a given
set of `targets`. The scraped metrics are forwarded to the list of receivers
passed in `forward_to`. Multiple `metrics.scrape` components can be specified
by providing a 'name' like "blackbox-scraper" in the following example.

## Example

The following example will set up the job with certain attributes (scrape
intervals, query parameters) and let it scrape two instances of the blackbox
exporter. The received metrics will be sent over to the provided list of
remote_writes, as defined by other components.

```river
metrics.scrape "blackbox-scraper" {
  targets = [
    {"__address__" = "blackbox-exporter:9115", "instance" = "one"},
    {"__address__" = "blackbox-exporter:9116", "instance" = "two"},
  ]
  forward_to = [metrics.remote_write.grafanacloud.receiver, metrics.remote_write.onprem.receiver]

  scrape_config {
    job_name        = "grafana"
    scrape_interval = "10s"
    params          = { "target" = ["grafana.com"], "module" = ["http_2xx"]}
    metrics_path    = "/probe"
  }
}
```

## Arguments
Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
targets   | list(map(string)) | The targets to scrape. | | **yes**
receivers | list(map(string)) | The targets to scrape. | | **yes**
extra_metrics| list(map(string)) | The targets to scrape. | false | no

### `scrape_config` block
The user must provide one `scrape_config` block; it will configure and start a
new scrape job to scrape all of the input targets. The list of arguments that
can be used to configure the block is presented below.

All arguments except for `job_name` are optional and any omitted fields will
take on their default values. In case that conflicting attributes are being
passed (eg. defining both a BearerToken and BearerTokenFile or configuring both
Basic Authorization and OAuth2 at the same time), the scrape job will not be
created.

Name                     | Type     | Description | Default | Required
------------------------ | -------- | ----------- | ------- | --------
job_name                 | string   | The job name to which the job label is set by default. | | **yes**
honor_labels             | bool     | Indicator whether the scraped metrics should remain unmodified. | false | no
honor_timestamps         | bool     | Indicator whether the scraped timestamps should be respected. | true | no
params                   | map(list(string)) | A set of query parameters with which the target is scraped. | | no
scrape_interval          | duration | How frequently to scrape the targets of this scrape config. | "60s" | no
scrape_timeout           | duration | The timeout for scraping targets of this config. | "10s" | no
metrics_path             | string   | The HTTP resource path on which to fetch metrics from targets. | /metrics | no
scheme                   | string   | The URL scheme with which to fetch metrics from targets. | | no
body_size_limit          | int      | An uncompressed response body larger than this many bytes will cause the scrape to fail. 0 means no limit. | | no
sample_limit             | uint     | More than this many samples post metric-relabeling will cause the scrape to fail | | no
target_limit             | uint     | More than this many targets after the target relabeling will cause the scrapes to fail. | | no
label_limit              | uint     | More than this many labels post metric-relabeling will cause the scrape to fail. | | no
label_name_length_limit  | uint     | More than this label name length post metric-relabeling will cause the | | no
label_value_length_limit | uint     | More than this label value length post metric-relabeling will cause the scrape to fail. | | no
basic_auth               | basic_auth block    | Setup of Basic HTTP authentication credentials. | | no
authorization            | authorization block | Setup of HTTP Authorization credentials. | | no
oauth2                   | oauth2 block        | Setup of the OAuth2 client. | | no
tls_config               | tls_config block    | Configuration options for TLS connections. | | no
bearer_token             | string   | Used to set up the Bearer Token. | | no
bearer_token_file        | string   | Used to set up the Bearer Token file. | | no
proxy_url                | string   | Used to set up a Proxy URL. | | no
follow_redirects         | bool     | Whether the scraper should follow redirects. | true | no
enable_http_2            | bool     | Whether the scraper should use HTTP2. | | no

#### basic_auth block
Name          | Type     | Description                                     | Default | Required
--------------| -------- | ----------------------------------------------- | ------- | -------
username      | string   | Setup of Basic HTTP authentication credentials. |         | no
password      | string   | Setup of Basic HTTP authentication credentials. |         | no
password_file | string   | Setup of Basic HTTP authentication credentials. |         | no

#### authorization block
Name                | Type     | Description                              | Default | Required
------------------- | -------- | -----------------------------------------| ------- | --------
type                | string   | Setup of HTTP Authorization credentials. |         | no
credential          | string   | Setup of HTTP Authorization credentials. |         | no
credentials_file    | string   | Setup of HTTP Authorization credentials. |         | no

#### oauth2 block
Name               | Type             | Description                              | Default | Required
------------------ | ---------------- | -----------------------------------------| ------- | --------
client_id          | string           | Setup of the OAuth2 client.              |         | no
client_secret      | string           | Setup of the OAuth2 client.              |         | no
client_secret_file | string           | Setup of the OAuth2 client.              |         | no
scopes             | list(string)     | Setup of the OAuth2 client.              |         | no
token_url          | string           | Setup of the OAuth2 client.              |         | no
endpoint_params    | map(string)      | Setup of the OAuth2 client.              |         | no
proxy_url          | string           | Setup of the OAuth2 client.              |         | no
tls_config         | tls_config block | Setup of TLS options.                    |         | no

#### tls_config block
Name                            | Type     | Description                                | Default | Required
------------------------------- | -------- | ------------------------------------------ | ------- | --------
tls_config_ca_file              | string   | Configuration options for TLS connections. |         | no
tls_config_cert_file            | string   | Configuration options for TLS connections. |         | no
tls_config_key_file             | string   | Configuration options for TLS connections. |         | no
tls_config_server_name          | string   | Configuration options for TLS connections. |         | no
tls_config_insecure_skip_verify | bool     | Configuration options for TLS connections. |         | no

## Exported fields
The `metrics.scrape` component does not export any fields that can be
referenced by other components.

## Component health
The `metrics.scrape` component will only be reported as unhealthy when
given an invalid configuration.

## Debug information
`metrics.scrape` reports the status of the last scrape for each configured
scrape job on the component's debug endpoint.

### Debug metrics
`targets.mutate` does not expose any component-specific debug metrics.

