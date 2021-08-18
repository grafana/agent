package crow

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
)

func Test_sample_Ready(t *testing.T) {
	tt := []struct {
		sample sample
		now    time.Time
		expect bool
	}{
		{
			sample: sample{
				ScrapeTime:        time.Unix(100, 0).UTC(),
				ValidationAttempt: 0,
			},
			now:    time.Unix(100, 0).UTC(),
			expect: false,
		},
		{
			sample: sample{
				ScrapeTime:        time.Unix(100, 0).UTC(),
				ValidationAttempt: 0,
			},
			now:    time.Unix(500, 0).UTC(),
			expect: true,
		},
	}

	for _, tc := range tt {
		ready := tc.sample.Ready(tc.now)
		require.Equal(t, tc.expect, ready)
	}
}

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

func Test_sampleGenerator(t *testing.T) {
	var (
		reg = prometheus.NewRegistry()
	)

	gen := sampleGenerator{
		numSamples: 10,
		sendCh:     make(chan<- []*sample, 10),
		r:          rand.New(rand.NewSource(0)),
	}
	reg.MustRegister(&gen)

	mfs, err := reg.Gather()
	require.NoError(t, err)

	var sb strings.Builder
	enc := expfmt.NewEncoder(&sb, expfmt.FmtText)
	for _, mf := range mfs {
		require.NoError(t, enc.Encode(mf))
	}

	expect := `# HELP crow_validation_sample Sample to validate
# TYPE crow_validation_sample gauge
crow_validation_sample{sample_num="sample_01"} 393152
crow_validation_sample{sample_num="sample_14"} 943416
crow_validation_sample{sample_num="sample_2f"} 980153
crow_validation_sample{sample_num="sample_51"} 637646
crow_validation_sample{sample_num="sample_55"} 976708
crow_validation_sample{sample_num="sample_94"} 995827
crow_validation_sample{sample_num="sample_c2"} 376202
crow_validation_sample{sample_num="sample_fa"} 126063
crow_validation_sample{sample_num="sample_fc"} 422456
crow_validation_sample{sample_num="sample_fd"} 197794
`
	require.Equal(t, expect, sb.String())
}
