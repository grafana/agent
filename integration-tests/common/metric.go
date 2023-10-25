package common

import (
	"encoding/json"
	"fmt"
)

type MetricResponse struct {
	Status string     `json:"status"`
	Data   MetricData `json:"data"`
}

type MetricData struct {
	ResultType string         `json:"resultType"`
	Result     []MetricResult `json:"result"`
}

type MetricResult struct {
	Metric Metric `json:"metric"`
	Value  Value  `json:"value"`
}

type Value struct {
	Timestamp int64
	Value     string
}

func (v *Value) UnmarshalJSON(b []byte) error {
	var arr []interface{}
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	if len(arr) != 2 {
		return fmt.Errorf("expected 2 values, got %d", len(arr))
	}
	v.Timestamp, _ = arr[0].(int64)
	v.Value, _ = arr[1].(string)
	return nil
}

type Metric struct {
	TestName string `json:"test_name"`
	Name     string `json:"__name__"`
}

func (m *MetricResponse) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}
