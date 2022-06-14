package transformer

import (
	"context"
	"fmt"

	"github.com/grafana/agent/component"
	"github.com/grafana/regexp"
	"github.com/hashicorp/hcl/v2"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/rfratto/gohcl"
)

func init() {
	component.Register(component.Registration{
		Name:    "discovery.transformer",
		Args:    Arguments{},
		Exports: Exports{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return New(opts, args.(Arguments))
		},
	})
}

// Arguments holds values which are used to configure the discovery.transformer component.
type Arguments struct {
	// Targets contains the input 'targets' passed by a service discovery component.
	Targets []Target `hcl:"targets,optional"`

	// The relabelling steps to apply to the each target's label set.
	RelabelConfigs []*RelabelConfig `hcl:"relabel_config,block"`
}

// Target refers to a singular HTTP or HTTPS endpoint that will be used for scraping.
// Here, we're using a map[string]string instead of labels.Labels; if the label ordering
// is important, we can change to follow the upstream logic instead.
// TODO (@tpaschalis) Maybe the target definitions should be part of the
// Service Discovery components package. Let's reconsider once it's ready.
type Target map[string]string

// RelabelConfig describes a relabelling step to be applied on a target.
type RelabelConfig struct {
	SourceLabels []string `hcl:"source_labels,optional"`
	Separator    string   `hcl:"separator,optional"`
	Regex        Regexp   `hcl:"regex,optional"`
	Modulus      uint64   `hcl:"modulus,optional"`
	TargetLabel  string   `hcl:"target_label,optional"`
	Replacement  string   `hcl:"replacement,optional"`
	Action       Action   `hcl:"action,optional"`
}

// DefaultRelabelConfig sets the default values of fields when decoding a RelabelConfig block.
var DefaultRelabelConfig = RelabelConfig{
	Action:      Replace,
	Separator:   ";",
	Regex:       mustNewRegexp("(.*)"),
	Replacement: "$1",
}

var relabelTarget = regexp.MustCompile(`^(?:(?:[a-zA-Z_]|\$(?:\{\w+\}|\w+))+\w*)+$`)

// DecodeHCL implements gohcl.Decoder.
// This method is only called on blocks, not objects.
func (rc *RelabelConfig) DecodeHCL(body hcl.Body, ctx *hcl.EvalContext) error {
	*rc = DefaultRelabelConfig

	type relabelConfig RelabelConfig
	err := gohcl.DecodeBody(body, ctx, (*relabelConfig)(rc))
	if err != nil {
		return err
	}

	if rc.Action == "" {
		return fmt.Errorf("relabel action cannot be empty")
	}
	if rc.Modulus == 0 && rc.Action == HashMod {
		return fmt.Errorf("relabel configuration for hashmod requires non-zero modulus")
	}
	if (rc.Action == Replace || rc.Action == HashMod || rc.Action == Lowercase || rc.Action == Uppercase) && rc.TargetLabel == "" {
		return fmt.Errorf("relabel configuration for %s action requires 'target_label' value", rc.Action)
	}
	if (rc.Action == Replace || rc.Action == Lowercase || rc.Action == Uppercase) && !relabelTarget.MatchString(rc.TargetLabel) {
		return fmt.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}
	if (rc.Action == Lowercase || rc.Action == Uppercase) && rc.Replacement != DefaultRelabelConfig.Replacement {
		return fmt.Errorf("'replacement' can not be set for %s action", rc.Action)
	}
	if rc.Action == LabelMap && !relabelTarget.MatchString(rc.Replacement) {
		return fmt.Errorf("%q is invalid 'replacement' for %s action", rc.Replacement, rc.Action)
	}
	if rc.Action == HashMod && !model.LabelName(rc.TargetLabel).IsValid() {
		return fmt.Errorf("%q is invalid 'target_label' for %s action", rc.TargetLabel, rc.Action)
	}

	if rc.Action == LabelDrop || rc.Action == LabelKeep {
		if rc.SourceLabels != nil ||
			rc.TargetLabel != DefaultRelabelConfig.TargetLabel ||
			rc.Modulus != DefaultRelabelConfig.Modulus ||
			rc.Separator != DefaultRelabelConfig.Separator ||
			rc.Replacement != DefaultRelabelConfig.Replacement {

			return fmt.Errorf("%s action requires only 'regex', and no other fields", rc.Action)
		}
	}

	return nil
}

// Exports holds values which are exported by the discovery.transformer component.
type Exports struct {
	Output []Target `hcl:"output,attr"`
}

// Component implements the discovery.transformer component.
type Component struct {
	opts component.Options
}

var (
	_ component.Component = (*Component)(nil)
)

// New creates a new discovery.transformer component.
func New(o component.Options, args Arguments) (*Component, error) {
	c := &Component{opts: o}

	// Call to Update() to set the output once at the start
	if err := c.Update(args); err != nil {
		return nil, err
	}

	return c, nil
}

// Run implements component.Component.
func (c *Component) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Update implements component.Component.
func (c *Component) Update(args component.Arguments) error {
	newArgs := args.(Arguments)

	targets := make([]Target, 0, len(newArgs.Targets))
	relabelConfigs := hclToPromRelabelConfigs(newArgs.RelabelConfigs)

	for _, t := range newArgs.Targets {
		lset := hclMapToPromLabels(t)
		lset = relabel.Process(lset, relabelConfigs...)
		if lset != nil {
			targets = append(targets, promLabelsToHCL(lset))
		}
	}

	c.opts.OnStateChange(Exports{
		Output: targets,
	})

	return nil
}

func hclMapToPromLabels(ls Target) labels.Labels {
	res := make([]labels.Label, 0, len(ls))
	for k, v := range ls {
		res = append(res, labels.Label{Name: k, Value: v})
	}

	return res
}

func promLabelsToHCL(ls labels.Labels) Target {
	res := make(map[string]string, len(ls))
	for _, l := range ls {
		res[l.Name] = l.Value
	}

	return res
}

func hclToPromRelabelConfigs(rcs []*RelabelConfig) []*relabel.Config {
	res := make([]*relabel.Config, len(rcs))
	for i, rc := range rcs {
		sourceLabels := make([]model.LabelName, len(rc.SourceLabels))
		for i, sl := range rc.SourceLabels {
			sourceLabels[i] = model.LabelName(sl)
		}

		res[i] = &relabel.Config{
			SourceLabels: sourceLabels,
			Separator:    rc.Separator,
			Modulus:      rc.Modulus,
			TargetLabel:  rc.TargetLabel,
			Replacement:  rc.Replacement,
			Action:       relabel.Action(rc.Action),
			Regex:        relabel.Regexp{Regexp: rc.Regex.Regexp}, // TODO (@tpaschalis) not super happy with how this turned out, let's check it again.
		}
	}

	return res
}
