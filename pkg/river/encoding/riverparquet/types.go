package riverparquet

import (
	"encoding/json"
)

type Row struct {
	ID         uint            `parquet:"id,delta"`
	ParentID   uint            `parquet:"parent_id,delta"`
	Name       string          `parquet:"name,dict"`
	Label      string          `parquet:"label,optional"`
	RiverType  string          `parquet:"river_type,dict"`
	RiverValue json.RawMessage `parquet:"river_value,json"`
}
