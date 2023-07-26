package stages

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/model"

	"golang.org/x/crypto/sha3"
)

// Config Errors.
var (
	ErrEmptyTemplateStageConfig = errors.New("template stage config cannot be empty")
	ErrTemplateSourceRequired   = errors.New("template source value is required")
)

var extraFunctionMap = template.FuncMap{
	"ToLower":    strings.ToLower,
	"ToUpper":    strings.ToUpper,
	"Replace":    strings.Replace,
	"Trim":       strings.Trim,
	"TrimLeft":   strings.TrimLeft,
	"TrimRight":  strings.TrimRight,
	"TrimPrefix": strings.TrimPrefix,
	"TrimSuffix": strings.TrimSuffix,
	"TrimSpace":  strings.TrimSpace,
	"Hash": func(salt string, input string) string {
		hash := sha3.Sum256([]byte(salt + input))
		return hex.EncodeToString(hash[:])
	},
	"Sha2Hash": func(salt string, input string) string {
		hash := sha256.Sum256([]byte(salt + input))
		return hex.EncodeToString(hash[:])
	},
	"regexReplaceAll": func(regex string, s string, repl string) string {
		r := regexp.MustCompile(regex)
		return r.ReplaceAllString(s, repl)
	},
	"regexReplaceAllLiteral": func(regex string, s string, repl string) string {
		r := regexp.MustCompile(regex)
		return r.ReplaceAllLiteralString(s, repl)
	},
}

var functionMap = sprig.TxtFuncMap()

func init() {
	for k, v := range extraFunctionMap {
		functionMap[k] = v
	}
}

// TemplateConfig configures template value extraction.
type TemplateConfig struct {
	Source   string `river:"source,attr"`
	Template string `river:"template,attr"`
}

// validateTemplateConfig validates the templateStage config.
func validateTemplateConfig(cfg TemplateConfig) (*template.Template, error) {
	if cfg.Source == "" {
		return nil, ErrTemplateSourceRequired
	}

	return template.New("pipeline_template").Funcs(functionMap).Parse(cfg.Template)
}

// newTemplateStage creates a new templateStage
func newTemplateStage(logger log.Logger, config TemplateConfig) (Stage, error) {
	t, err := validateTemplateConfig(config)
	if err != nil {
		return nil, err
	}

	return toStage(&templateStage{
		cfgs:     config,
		logger:   logger,
		template: t,
	}), nil
}

// templateStage will mutate the incoming entry and set it from extracted data
type templateStage struct {
	cfgs     TemplateConfig
	logger   log.Logger
	template *template.Template
}

// Process implements Stage
func (o *templateStage) Process(labels model.LabelSet, extracted map[string]interface{}, t *time.Time, entry *string) {
	td := make(map[string]interface{})
	for k, v := range extracted {
		s, err := getString(v)
		if err != nil {
			if Debug {
				level.Debug(o.logger).Log("msg", "extracted template could not be converted to a string", "err", err, "type", reflect.TypeOf(v))
			}
			continue
		}
		td[k] = s
		if k == o.cfgs.Source {
			td["Value"] = s
		}
	}
	td["Entry"] = *entry

	buf := &bytes.Buffer{}
	err := o.template.Execute(buf, td)
	if err != nil {
		if Debug {
			level.Debug(o.logger).Log("msg", "failed to execute template on extracted value", "err", err)
		}
		return
	}
	st := buf.String()
	// If the template evaluates to an empty string, remove the key from the map
	if st == "" {
		delete(extracted, o.cfgs.Source)
	} else {
		extracted[o.cfgs.Source] = st
	}
}

// Name implements Stage
func (o *templateStage) Name() string {
	return StageTypeTemplate
}
