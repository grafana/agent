package flow

import (
	"github.com/grafana/agent/pkg/flow/internal/funcs"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

var rootEvalContext = &hcl.EvalContext{
	// NOTE(rfratto): Terraform doesn't delimit multiple words in function names,
	// but we use snake_case.

	Functions: map[string]function.Function{
		"abs":           stdlib.AbsoluteFunc,
		"ceil":          stdlib.CeilFunc,
		"chomp":         stdlib.ChompFunc,
		"coalesce_list": stdlib.CoalesceListFunc,
		"compact":       stdlib.CompactFunc,
		"concat":        stdlib.ConcatFunc,
		"contains":      stdlib.ContainsFunc,
		"csv_decode":    stdlib.CSVDecodeFunc,
		"distinct":      stdlib.DistinctFunc,
		"element":       stdlib.ElementFunc,
		"env":           funcs.EnvFunc,
		"chunk_list":    stdlib.ChunklistFunc,
		"flatten":       stdlib.FlattenFunc,
		"floor":         stdlib.FloorFunc,
		"format":        stdlib.FormatFunc,
		"format_date":   stdlib.FormatDateFunc,
		"format_list":   stdlib.FormatListFunc,
		"indent":        stdlib.IndentFunc,
		// TODO(rfratto): stdlib's IndexFunc and Terraform's IndexFunc are
		// incompatible; we probably want to emulate the latter to not surprise
		// people.
		"join":             stdlib.JoinFunc,
		"json_decode":      stdlib.JSONDecodeFunc,
		"json_encode":      stdlib.JSONEncodeFunc,
		"keys":             stdlib.KeysFunc,
		"log":              stdlib.LogFunc,
		"lower":            stdlib.LowerFunc,
		"max":              stdlib.MaxFunc,
		"merge":            stdlib.MergeFunc,
		"min":              stdlib.MinFunc,
		"parse_int":        stdlib.ParseIntFunc,
		"pow":              stdlib.PowFunc,
		"range":            stdlib.RangeFunc,
		"regex":            stdlib.RegexFunc,
		"regex_all":        stdlib.RegexAllFunc,
		"reverse":          stdlib.ReverseListFunc,
		"set_intersection": stdlib.SetIntersectionFunc,
		"set_product":      stdlib.SetProductFunc,
		"set_subtract":     stdlib.SetSubtractFunc,
		"set_union":        stdlib.SetUnionFunc,
		"number_sign":      stdlib.SignumFunc,
		"slice":            stdlib.SliceFunc,
		"sort":             stdlib.SortFunc,
		"split":            stdlib.SplitFunc,
		"string_reverse":   stdlib.ReverseFunc,
		"substr":           stdlib.SubstrFunc,
		"time_add":         stdlib.TimeAddFunc,
		"title":            stdlib.TitleFunc,
		"trim":             stdlib.TrimFunc,
		"trim_prefix":      stdlib.TrimPrefixFunc,
		"trim_space":       stdlib.TrimSpaceFunc,
		"trim_suffix":      stdlib.TrimSuffixFunc,
		"upper":            stdlib.UpperFunc,
		"values":           stdlib.ValuesFunc,
		"zipmap":           stdlib.ZipmapFunc,
	},
}
