// TODO: Does this file have to exist? Should we move its contents elsewhere?
package observer

import (
	"bytes"
	"time"

	"github.com/grafana/agent/component"
	"github.com/grafana/agent/pkg/river/encoding/riverparquet"
	"github.com/parquet-go/parquet-go"
)

// createParquet creates the parquet file out of agent state structures.
func createParquet(metadata map[string]string, rows []componentRow) ([]byte, error) {
	var buf bytes.Buffer
	writer := parquet.NewGenericWriter[componentRow](&buf)

	// Write the component data to the buffer.
	rowGroup := parquet.NewGenericBuffer[componentRow]()
	_, err := rowGroup.Write(rows)
	if err != nil {
		return nil, err
	}

	_, err = writer.WriteRowGroup(rowGroup)
	if err != nil {
		return nil, err
	}

	// Write the metadata to the buffer.
	for key, label := range metadata {
		writer.SetKeyValueMetadata(key, label)
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func convertToParquetRows(components []*component.Info) []componentRow {
	res := []componentRow{}

	for _, cInfo := range components {
		var (
			args      = riverparquet.GetComponentDetail(cInfo.Arguments)
			exports   = riverparquet.GetComponentDetail(cInfo.Exports)
			debugInfo = riverparquet.GetComponentDetail(cInfo.DebugInfo)
		)

		componentState := componentRow{
			ID:       cInfo.ID.LocalID,
			ModuleID: cInfo.ID.ModuleID,
			Health: componentHealth{
				Health:     cInfo.Health.Health.String(),
				Message:    cInfo.Health.Message,
				UpdateTime: cInfo.Health.UpdateTime,
			},
			Arguments: args,
			Exports:   exports,
			DebugInfo: debugInfo,
		}

		res = append(res, componentState)
	}

	return res
}

type componentRow struct {
	ID        string             `parquet:"id"`
	ModuleID  string             `parquet:"module_id"`
	Health    componentHealth    `parquet:"health"`
	Arguments []riverparquet.Row `parquet:"arguments"`
	Exports   []riverparquet.Row `parquet:"exports"`
	DebugInfo []riverparquet.Row `parquet:"debug_info"`
}

type componentHealth struct {
	Health     string    `parquet:"state"`
	Message    string    `parquet:"message"`
	UpdateTime time.Time `parquet:"update_time"`
}
