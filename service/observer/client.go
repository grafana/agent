// TODO: Does this file have to exist? Should we move its contents elsewhere?
package observer

import (
	"bytes"
	"context"

	"github.com/parquet-go/parquet-go"
)

type AgentStateWriter interface {
	Write(ctx context.Context, agentState []byte) error
}

// GetAgentStateParquet creates the parquet file out of agent state structures.
func GetAgentStateParquet(labels map[string]string, components []componentRow) ([]byte, error) {
	var buf bytes.Buffer
	writer := parquet.NewGenericWriter[componentRow](&buf)

	// Write the component data to the buffer.
	rowGroup := parquet.NewGenericBuffer[componentRow]()
	_, err := rowGroup.Write(components)
	if err != nil {
		return nil, err
	}

	_, err = writer.WriteRowGroup(rowGroup)
	if err != nil {
		return nil, err
	}

	// Write the metadata to the buffer.
	for key, label := range labels {
		writer.SetKeyValueMetadata(key, label)
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
