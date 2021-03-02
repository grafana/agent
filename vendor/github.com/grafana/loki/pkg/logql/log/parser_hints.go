package log

import (
	"strings"
)

var noParserHints = &parserHint{}

// ParserHint are hints given to LogQL parsers.
// This is specially useful for parser that extract implicitly all possible label keys.
// This is used only within metric queries since it's rare that you need all label keys.
// For example in the following expression:
//
//		sum by (status_code) (rate({app="foo"} | json [5m]))
//
// All we need to extract is the status_code in the json parser.
type ParserHint interface {
	// Tells if a label with the given key should be extracted.
	ShouldExtract(key string) bool
	// Tells if there's any hint that start with the given prefix.
	// This allows to speed up key searching in nested structured like json.
	ShouldExtractPrefix(prefix string) bool
	// Tells if we should not extract any labels.
	// For example in :
	//		 sum(rate({app="foo"} | json [5m]))
	// We don't need to extract any labels from the log line.
	NoLabels() bool
}

type parserHint struct {
	noLabels       bool
	requiredLabels []string
}

func (p *parserHint) ShouldExtract(key string) bool {
	if len(p.requiredLabels) == 0 {
		return true
	}
	for _, l := range p.requiredLabels {
		if l == key {
			return true
		}
	}
	return false
}

func (p *parserHint) ShouldExtractPrefix(prefix string) bool {
	if len(p.requiredLabels) == 0 {
		return true
	}
	for _, l := range p.requiredLabels {
		if strings.HasPrefix(l, prefix) {
			return true
		}
	}

	return false
}

func (p *parserHint) NoLabels() bool {
	return p.noLabels
}

// newParserHint creates a new parser hint using the list of labels that are seen and required in a query.
func newParserHint(requiredLabelNames, groups []string, without, noLabels bool, metricLabelName string) *parserHint {
	if len(groups) > 0 {
		requiredLabelNames = append(requiredLabelNames, groups...)
	}
	if metricLabelName != "" {
		requiredLabelNames = append(requiredLabelNames, metricLabelName)
	}
	requiredLabelNames = uniqueString(requiredLabelNames)
	if noLabels {
		if len(requiredLabelNames) > 0 {
			return &parserHint{requiredLabels: requiredLabelNames}
		}
		return &parserHint{noLabels: true}
	}
	// we don't know what is required when a without clause is used.
	// Same is true when there's no grouping.
	// no hints available then.
	if without || len(groups) == 0 {
		return noParserHints
	}
	return &parserHint{requiredLabels: requiredLabelNames}
}
