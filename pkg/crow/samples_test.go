package crow

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_sampleBackoff(t *testing.T) {
	tt := []struct {
		attempt int
		expect  time.Duration
	}{
		{attempt: 0, expect: 1250 * time.Millisecond},
		{attempt: 1, expect: 1500 * time.Millisecond},
		{attempt: 2, expect: 2000 * time.Millisecond},
		{attempt: 3, expect: 3000 * time.Millisecond},
		{attempt: 4, expect: 5000 * time.Millisecond},
		{attempt: 5, expect: 9000 * time.Millisecond},
	}

	for _, tc := range tt {
		t.Run(fmt.Sprintf("%d", tc.attempt), func(t *testing.T) {
			actual := sampleBackoff(tc.attempt)
			require.Equal(t, tc.expect, actual)
		})
	}
}
