package statsd

import (
	"testing"
	"time"

	"github.com/grafana/agent/pkg/river"
	"github.com/stretchr/testify/require"
)

var (
	exampleRiverConfig = `
		listen_udp						= "1010"
		listen_tcp						= "1011"
		listen_unixgram					= "unix"
		unix_socket_mode				= "prom"
		mapping_config_path 			= "mapTest.yaml"
		read_buffer						= 1
		cache_size						= 2
		cache_type						= "random"
		event_queue_size				= 1000
		event_flush_interval			= "1m"
		parse_dogstatsd_tags			= true
		parse_influxdb_tags				= false
		parse_librato_tags				= false
		parse_signalfx_tags				= false
		`
	duration1s, _ = time.ParseDuration("1s")
	duration1m, _ = time.ParseDuration("1m")
)

func TestRiverUnmarshall(t *testing.T) {

	var args Config
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	require.Equal(t, "1010", string(args.ListenUDP))
	require.Equal(t, "1011", string(args.ListenTCP))
	require.Equal(t, "unix", string(args.ListenUnixgram))
	require.Equal(t, "prom", string(args.UnixSocketMode))
	require.Equal(t, 1, int(args.ReadBuffer))
	require.Equal(t, 2, int(args.CacheSize))
	require.Equal(t, "random", string(args.CacheType))
	require.Equal(t, 1000, int(args.EventQueueSize))
	require.Equal(t, duration1m, args.EventFlushInterval)
	require.Equal(t, true, args.ParseDogStatsd)
	require.Equal(t, false, args.ParseInfluxDB)
	require.Equal(t, false, args.ParseLibrato)
	require.Equal(t, false, args.ParseSignalFX)
	require.Equal(t, "mapTest.yaml", args.MappingConfig)
}

func TestConvert(t *testing.T) {
	var args Config
	err := river.Unmarshal([]byte(exampleRiverConfig), &args)
	require.NoError(t, err)

	configStatsd, err2 := args.Convert()
	require.NoError(t, err2)

	require.Equal(t, "1010", string(configStatsd.ListenUDP))
	require.Equal(t, "1011", string(configStatsd.ListenTCP))
	require.Equal(t, "unix", string(configStatsd.ListenUnixgram))
	require.Equal(t, "prom", string(configStatsd.UnixSocketMode))
	require.Equal(t, 1, int(configStatsd.ReadBuffer))
	require.Equal(t, 2, int(configStatsd.CacheSize))
	require.Equal(t, "random", string(configStatsd.CacheType))
	require.Equal(t, 1000, int(configStatsd.EventQueueSize))
	require.Equal(t, duration1m, configStatsd.EventFlushInterval)
	require.Equal(t, true, configStatsd.ParseDogStatsd)
	require.Equal(t, false, configStatsd.ParseInfluxDB)
	require.Equal(t, false, configStatsd.ParseLibrato)
	require.Equal(t, false, configStatsd.ParseSignalFX)
}
