package schema

import (
	"fmt"
	"github.com/grafana/agent/component/prometheus/scrape"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type TestArguments struct {
	SimpleAttr      string `river:"simple_attr,attr"`
	OptionalAttr    string `river:"optional_attr,attr,optional"`
	notUsed         int
	OptionalAttrPtr *string `river:"optional_attr_ptr,attr,optional"`

	SimpleArray    []string  `river:"simple_arr,attr"`
	SimpleArrayOpt []string  `river:"simple_arr_opt,attr,optional"`
	SimpleArrayPtr *[]string `river:"simple_arr_ptr,attr,optional"`

	SimpleMap    map[string]string `river:"simple_map,attr"`
	SimpleMapOpt map[string]string `river:"simple_map_opt,attr,optional"`

	SimpleDuration    time.Duration  `river:"simple_duration,attr"`
	SimpleDurationOpt time.Duration  `river:"simple_duration_opt,attr,optional"`
	SimpleDurationPtr *time.Duration `river:"simple_duration_ptr,attr,optional"`

	AttrObject    Block  `river:"attr_object,attr"`
	AttrObjectOpt Block  `river:"attr_object_opt,attr,optional"`
	AttrObjectPtr *Block `river:"attr_object_ptr,attr,optional"`

	otherNotUsed otherNotUsed

	Block    Block  `river:"block,block"`
	BlockOpt Block  `river:"block_opt,block,optional"`
	BlockPtr *Block `river:"block_ptr,block,optional"`
}

type Block struct {
	SimpleAttr      int     `river:"simple,attr"`
	OptionalAttr    int     `river:"optional_attr,attr,optional"`
	OptionalAttrPtr *string `river:"optional_ptr,attr,optional"`
}

type otherNotUsed struct {
	NotUsed int
}

func TestRiverToYAML_Example(t *testing.T) {
	expectedYAML := `- type: string
  name: simple_attr
  flags:
  - attr
- type: string
  name: optional_attr
  flags:
  - attr
  - optional
- type: string
  name: optional_attr_ptr
  flags:
  - attr
  - optional
- type: list(string)
  name: simple_arr
  flags:
  - attr
- type: list(string)
  name: simple_arr_opt
  flags:
  - attr
  - optional
- type: list(string)
  name: simple_arr_ptr
  flags:
  - attr
  - optional
- type: map(string)
  name: simple_map
  flags:
  - attr
- type: map(string)
  name: simple_map_opt
  flags:
  - attr
  - optional
- type: duration
  name: simple_duration
  flags:
  - attr
- type: duration
  name: simple_duration_opt
  flags:
  - attr
  - optional
- type: duration
  name: simple_duration_ptr
  flags:
  - attr
  - optional
- type: Block
  name: attr_object
  flags:
  - attr
  nested:
  - type: int
    name: simple
    flags:
    - attr
  - type: int
    name: optional_attr
    flags:
    - attr
    - optional
  - type: string
    name: optional_ptr
    flags:
    - attr
    - optional
- type: Block
  name: attr_object_opt
  flags:
  - attr
  - optional
  nested:
  - type: int
    name: simple
    flags:
    - attr
  - type: int
    name: optional_attr
    flags:
    - attr
    - optional
  - type: string
    name: optional_ptr
    flags:
    - attr
    - optional
- type: Block
  name: attr_object_ptr
  flags:
  - attr
  - optional
  nested:
  - type: int
    name: simple
    flags:
    - attr
  - type: int
    name: optional_attr
    flags:
    - attr
    - optional
  - type: string
    name: optional_ptr
    flags:
    - attr
    - optional
- type: Block
  name: block
  flags:
  - block
  nested:
  - type: int
    name: simple
    flags:
    - attr
  - type: int
    name: optional_attr
    flags:
    - attr
    - optional
  - type: string
    name: optional_ptr
    flags:
    - attr
    - optional
- type: Block
  name: block_opt
  flags:
  - block
  - optional
  nested:
  - type: int
    name: simple
    flags:
    - attr
  - type: int
    name: optional_attr
    flags:
    - attr
    - optional
  - type: string
    name: optional_ptr
    flags:
    - attr
    - optional
- type: Block
  name: block_ptr
  flags:
  - block
  - optional
  nested:
  - type: int
    name: simple
    flags:
    - attr
  - type: int
    name: optional_attr
    flags:
    - attr
    - optional
  - type: string
    name: optional_ptr
    flags:
    - attr
    - optional
`
	yaml, err := RiverToYAML(TestArguments{})
	fmt.Println(yaml)
	assert.NoError(t, err)
	assert.Equal(t, expectedYAML, yaml)

	yaml, err = RiverToYAML(&TestArguments{})
	assert.NoError(t, err)
	assert.Equal(t, expectedYAML, yaml)
}

func TestRiverToYAML_PrometheusScrapeArguments(t *testing.T) {
	expectedYAML := `- type: list(map(string))
  name: targets
  flags:
  - attr
- type: list(Appendable)
  name: forward_to
  flags:
  - attr
- type: string
  name: job_name
  flags:
  - attr
  - optional
- type: bool
  name: honor_labels
  flags:
  - attr
  - optional
- type: bool
  name: honor_timestamps
  flags:
  - attr
  - optional
- type: map(list(string))
  name: params
  flags:
  - attr
  - optional
- type: bool
  name: scrape_classic_histograms
  flags:
  - attr
  - optional
- type: duration
  name: scrape_interval
  flags:
  - attr
  - optional
- type: duration
  name: scrape_timeout
  flags:
  - attr
  - optional
- type: string
  name: metrics_path
  flags:
  - attr
  - optional
- type: string
  name: scheme
  flags:
  - attr
  - optional
- type: Base2Bytes
  name: body_size_limit
  flags:
  - attr
  - optional
- type: uint
  name: sample_limit
  flags:
  - attr
  - optional
- type: uint
  name: target_limit
  flags:
  - attr
  - optional
- type: uint
  name: label_limit
  flags:
  - attr
  - optional
- type: uint
  name: label_name_length_limit
  flags:
  - attr
  - optional
- type: uint
  name: label_value_length_limit
  flags:
  - attr
  - optional
- type: BasicAuth
  name: basic_auth
  flags:
  - block
  - optional
  nested:
  - type: string
    name: username
    flags:
    - attr
    - optional
  - type: Secret
    name: password
    flags:
    - attr
    - optional
  - type: string
    name: password_file
    flags:
    - attr
    - optional
- type: Authorization
  name: authorization
  flags:
  - block
  - optional
  nested:
  - type: string
    name: type
    flags:
    - attr
    - optional
  - type: Secret
    name: credentials
    flags:
    - attr
    - optional
  - type: string
    name: credentials_file
    flags:
    - attr
    - optional
- type: OAuth2Config
  name: oauth2
  flags:
  - block
  - optional
  nested:
  - type: string
    name: client_id
    flags:
    - attr
    - optional
  - type: Secret
    name: client_secret
    flags:
    - attr
    - optional
  - type: string
    name: client_secret_file
    flags:
    - attr
    - optional
  - type: list(string)
    name: scopes
    flags:
    - attr
    - optional
  - type: string
    name: token_url
    flags:
    - attr
    - optional
  - type: map(string)
    name: endpoint_params
    flags:
    - attr
    - optional
  - type: URL
    name: proxy_url
    flags:
    - attr
    - optional
  - type: TLSConfig
    name: tls_config
    flags:
    - block
    - optional
    nested:
    - type: string
      name: ca_pem
      flags:
      - attr
      - optional
    - type: string
      name: ca_file
      flags:
      - attr
      - optional
    - type: string
      name: cert_pem
      flags:
      - attr
      - optional
    - type: string
      name: cert_file
      flags:
      - attr
      - optional
    - type: Secret
      name: key_pem
      flags:
      - attr
      - optional
    - type: string
      name: key_file
      flags:
      - attr
      - optional
    - type: string
      name: server_name
      flags:
      - attr
      - optional
    - type: bool
      name: insecure_skip_verify
      flags:
      - attr
      - optional
    - type: TLSVersion
      name: min_version
      flags:
      - attr
      - optional
- type: Secret
  name: bearer_token
  flags:
  - attr
  - optional
- type: string
  name: bearer_token_file
  flags:
  - attr
  - optional
- type: URL
  name: proxy_url
  flags:
  - attr
  - optional
- type: TLSConfig
  name: tls_config
  flags:
  - block
  - optional
  nested:
  - type: string
    name: ca_pem
    flags:
    - attr
    - optional
  - type: string
    name: ca_file
    flags:
    - attr
    - optional
  - type: string
    name: cert_pem
    flags:
    - attr
    - optional
  - type: string
    name: cert_file
    flags:
    - attr
    - optional
  - type: Secret
    name: key_pem
    flags:
    - attr
    - optional
  - type: string
    name: key_file
    flags:
    - attr
    - optional
  - type: string
    name: server_name
    flags:
    - attr
    - optional
  - type: bool
    name: insecure_skip_verify
    flags:
    - attr
    - optional
  - type: TLSVersion
    name: min_version
    flags:
    - attr
    - optional
- type: bool
  name: follow_redirects
  flags:
  - attr
  - optional
- type: bool
  name: enable_http2
  flags:
  - attr
  - optional
- type: bool
  name: extra_metrics
  flags:
  - attr
  - optional
- type: bool
  name: enable_protobuf_negotiation
  flags:
  - attr
  - optional
- type: ComponentBlock
  name: clustering
  flags:
  - block
  - optional
  nested:
  - type: bool
    name: enabled
    flags:
    - attr
`
	yaml, err := RiverToYAML(scrape.Arguments{})
	fmt.Println(yaml)
	assert.NoError(t, err)
	assert.Equal(t, expectedYAML, yaml)
}
