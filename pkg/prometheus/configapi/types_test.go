package configapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContentTypeFromRequest(t *testing.T) {
	tt := []struct {
		name        string
		contentType string
		expected    ContentType
	}{
		{"json", "application/json", ContentTypeJSON},
		{"yaml", "text/yaml", ContentTypeYAML},
		{"default", "", DefaultContentType},
		{"invalid", "application/xml", ContentTypeUnknown},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			headers := http.Header{
				"Content-Type": []string{tc.contentType},
			}
			r := http.Request{Header: headers}
			out, _ := ContentTypeFromRequest(&r)

			require.Equal(t, tc.expected, out)
		})
	}
}
