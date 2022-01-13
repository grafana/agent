package configapi

import (
	"encoding/json"
	"fmt"
	"io"
)

// UnmarshalAPIResponse unmarshals a JSON APIResponse from r. The "data" field
// of the API response will be unmarshaled into v. If the response was a failure,
// the failure will be returned as an instance of Error.
func UnmarshalAPIResponse(r io.Reader, v interface{}) error {
	type rawResponse struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	var resp rawResponse

	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&resp); err != nil {
		return fmt.Errorf("could not read response: %w", err)
	}

	switch resp.Status {
	case "success":
		// Nothing to decode into. Success.
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(resp.Data, v); err != nil {
			return fmt.Errorf("unmarshaling data: %w", err)
		}
		return nil

	case "error":
		var errResp ErrorResponse
		if err := json.Unmarshal(resp.Data, &errResp); err != nil {
			return fmt.Errorf("unmarshaling error: %w", err)
		}
		return Error{Message: errResp.Error}

	default:
		return fmt.Errorf("unknown API response status %q", resp.Status)
	}
}

// Error is the error returned by UnmarshalAPIResponse when the unmarshal was
// successful but the API itself indicated an error occurred.
type Error struct {
	Message string
}

// Error implements error.
func (e Error) Error() string { return e.Message }
