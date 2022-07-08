# metrics.scrape
The `metrics.scrape` component configures scrape jobs for scraping metrics from
a given set of `targets`. The metrics are forwarded to the given list of
`receivers`. Multiple `metrics.scrape` components can be specified by
providing a 'name' like "blackbox-scraper" in the following example.

## Example

The following example will set up two scrape jobs with different attributes
(scrape intervals, query parameters) and let them scrape the two instances of
the blackbox exporter. The received metrics will be forwarded to the provided
remote_write 'receivers' which are referred to as exports from another
component.

```hcl
metrics "scrape" "blackbox-scraper" {
  targets = [ 
    {"__address__" = "blackbox-exporter:9115", "instance" = "one"},
    {"__address__" = "blackbox-exporter:9116", "instance" = "two"},
  ]
  receivers = [metrics.remote_write.grafanacloud, metrics.remote_write.onprem]

  scrape_config {
    job_name        = "grafana"
    scrape_interval = "10s"
    params          = { "target" = ["grafana.com"], "module" = ["http_2xx"]}
    metrics_path    = "/probe"
  }

  scrape_config {
    job_name        = "google"
    scrape_interval = "120s"
    params          = { "target" = ["google.com"], "module" = ["http_2xx"]}
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
The user can provide zero or more `scrape_config` blocks; each one will
configure and start a new scrape job to scrape all of the input targets. The
list of arguments that can be used to configure a block is presented below.

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
metrics_path             | string   | The HTTP resource path on which to fetch metrics from targets. | /metrics | 
scheme                   | string   | The URL scheme with which to fetch metrics from targets. | | 
body_size_limit          | int      | An uncompressed response body larger than this many bytes will cause the scrape to fail. 0 means no limit. | | no 
sample_limit             | uint     | More than this many samples post metric-relabeling will cause the scrape to fail | | no 
target_limit             | uint     | More than this many targets after the target relabeling will cause the scrapes to fail. | | no 
label_limit              | uint     | More than this many labels post metric-relabeling will cause the scrape to fail. | | no 
label_name_length_limit  | uint     | More than this label name length post metric-relabeling will cause the | | no 
label_value_length_limit | uint     | More than this label value length post metric-relabeling will cause the scrape to fail. | | no 
basic_auth_username      | string   | Setup of Basic HTTP authentication credentials. | | 
basic_auth_password      | string   | Setup of Basic HTTP authentication credentials. | |  
basic_auth_password_file | string   | Setup of Basic HTTP authentication credentials. | | 
authorization_type       | string   | Setup of HTTP Authorization credentials. | | 
authorization_credential | string   | Setup of HTTP Authorization credentials. | | 
authorization_credentials_file    | string | Setup of HTTP Authorization credentials. | | 
oauth2_client_id         | string   | Setup of the OAuth2 client. | | 
oauth2_client_secret     | string   | Setup of the OAuth2 client. | | no 
oauth2_client_secret_file | string  | Setup of the OAuth2 client. | | 
oauth2_scopes            | list(string) | Setup of the OAuth2 client. | | no 
oauth2_token_url         | string   | Setup of the OAuth2 client. | | no 
oauth2_endpoint_params   | map(string) | Setup of the OAuth2 client. | | no 
oauth2_proxy_url         | string   | Setup of the OAuth2 client. | | no 
oauth2_tls_config_ca_file     | string | Setup of the OAuth2 client. | | no 
oauth2_tls_config_cert_file   | string | Setup of the OAuth2 client. | | no 
oauth2_tls_config_key_file    | string | Setup of the OAuth2 client. | | no 
oauth2_tls_config_server_name | string | Setup of the OAuth2 client. | | no 
oauth2_tls_config_insecure_skip_verify    | bool | Setup of the OAuth2 client. | | no 
bearer_token             | string   | Used to set up the Bearer Token | | no 
bearer_token_file        | string   | Used to set up the Bearer Token file | | no 
proxy_url                | string   | Used to set up a Proxy URL | | no 
tls_config_ca_file       | string   | Configuration options for TLS connections | | no 
tls_config_cert_file     | string   | Configuration options for TLS connections | | no 
tls_config_key_file      | string   | Configuration options for TLS connections | | no 
tls_config_server_name   | string   | Configuration options for TLS connections | | no 
tls_config_insecure_skip_verify | bool | Configuration options for TLS connections | | no 
follow_redirects         | bool     | Whether the scraper should follow redirects | true | no 
enable_http_2            | bool     | Whether the scraper should use HTTP2 | | no 

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

