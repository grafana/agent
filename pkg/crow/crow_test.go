package crow

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
)

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
crow_validation_sample{sample_num="sample_0"} 165505
crow_validation_sample{sample_num="sample_1"} 393152
crow_validation_sample{sample_num="sample_2"} 995827
crow_validation_sample{sample_num="sample_3"} 197794
crow_validation_sample{sample_num="sample_4"} 376202
crow_validation_sample{sample_num="sample_5"} 126063
crow_validation_sample{sample_num="sample_6"} 980153
crow_validation_sample{sample_num="sample_7"} 422456
crow_validation_sample{sample_num="sample_8"} 894929
crow_validation_sample{sample_num="sample_9"} 637646
`
	require.Equal(t, expect, sb.String())
}
