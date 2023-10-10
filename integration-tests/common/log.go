package common

import "encoding/json"

type LogResponse struct {
	Status string    `json:"status"`
	Data   []LogData `json:"data"`
}

type LogData struct {
	Filename string `json:"filename"`
}

func (m *LogResponse) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}
