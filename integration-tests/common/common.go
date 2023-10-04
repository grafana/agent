package common

import (
	"errors"
	"io"
	"net/http"
	"time"
)

type Unmarshaler interface {
	Unmarshal([]byte) error
}

const DefaultRetryInterval = time.Second * 5
const DefaultTimeout = time.Minute * 3

func FetchDataFromURL(url string, target Unmarshaler) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Non-OK HTTP status: " + resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return target.Unmarshal(bodyBytes)
}
