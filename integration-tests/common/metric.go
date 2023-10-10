package common

import "encoding/json"

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
}

type Metric struct {
	Job  string `json:"job"`
	Name string `json:"__name__"`
}

func (m *MetricResponse) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}
