package transformer

import (
	"fmt"

	"github.com/grafana/regexp"
)

// Action is the relabelling action to be performed.
type Action string

// Regexp encapsulates the Regexp type from Grafana's
// fork of the Go stdlib regexp package.
// TODO (@tpaschalis) This encapsulation already exists in Prometheus' relabel.Regexp
// so not sure whether to also move it here for now.
type Regexp struct {
	re       *regexp.Regexp
	original string
}

// All possible Action values.
const (
	Replace   Action = "replace"
	Keep      Action = "keep"
	Drop      Action = "drop"
	HashMod   Action = "hashmod"
	LabelMap  Action = "labelmap"
	LabelDrop Action = "labeldrop"
	LabelKeep Action = "labelkeep"
	Lowercase Action = "lowercase"
	Uppercase Action = "uppercase"
)

var actions = map[Action]struct{}{
	Replace:   {},
	Keep:      {},
	Drop:      {},
	HashMod:   {},
	LabelMap:  {},
	LabelDrop: {},
	LabelKeep: {},
	Lowercase: {},
	Uppercase: {},
}

// String returns the string representation of the Action type.
func (a Action) String() string {
	if _, exists := actions[a]; exists {
		return string(a)
	}
	return "Action:" + string(a)
}

// MarshalText implements encoding.TextMarshaler for Action.
func (a Action) MarshalText() (text []byte, err error) {
	return []byte(a.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for Action.
func (a *Action) UnmarshalText(text []byte) error {
	if _, exists := actions[Action(text)]; exists {
		*a = Action(text)
		return nil
	}
	return fmt.Errorf("unrecognized action type %s", string(text))
}

// NewRegexp creates a new anchored Regexp and returns an error if the
// passed-in regular expression does not compile.
func NewRegexp(s string) (Regexp, error) {
	regex, err := regexp.Compile("^(?:" + s + ")$")
	return Regexp{
		re:       regex,
		original: s,
	}, err
}

// MustNewRegexp works like NewRegexp, but panics if the regular expression does not compile.
func MustNewRegexp(s string) Regexp {
	re, err := NewRegexp(s)
	if err != nil {
		panic(err)
	}
	return re
}

// MarshalText implements encoding.TextMarshaler for Regexp.
func (re Regexp) MarshalText() (text []byte, err error) {
	if re.original != "" {
		return []byte(re.original), nil
	}
	return nil, nil
}

// UnmarshalText implements encoding.TextUnmarshaler for Regexp.
func (re *Regexp) UnmarshalText(text []byte) error {
	regex, err := regexp.Compile("^(?:" + string(text) + ")$")
	if err != nil {
		return err
	}

	re = &Regexp{
		re:       regex,
		original: string(text),
	}
	return nil
}
