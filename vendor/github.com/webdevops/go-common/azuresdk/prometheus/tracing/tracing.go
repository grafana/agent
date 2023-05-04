package tracing

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	HostnameMaxParts = 3

	EnvVarApiRequestBuckets     = "METRIC_AZURERM_API_REQUEST_BUCKETS"
	EnvVarApiRequestEnabled     = "METRIC_AZURERM_API_REQUEST_ENABLE"
	EnvVarApiRequestLables      = "METRIC_AZURERM_API_REQUEST_LABELS"
	EnvVarApiRatelimitEnabled   = "METRIC_AZURERM_API_RATELIMIT_ENABLE"
	EnvVarApiRatelimitAutoreset = "METRIC_AZURERM_API_RATELIMIT_AUTORESET"
)

type (
	// azure jwt auth token from header
	azureJwtAuthToken struct {
		Aud      string   `json:"aud"`
		Iss      string   `json:"iss"`
		Iat      int      `json:"iat"`
		Nbf      int      `json:"nbf"`
		Exp      int      `json:"exp"`
		Aio      string   `json:"aio"`
		Appid    string   `json:"appid"`
		Appidacr string   `json:"appidacr"`
		Groups   []string `json:"groups"`
		Idp      string   `json:"idp"`
		Idtyp    string   `json:"idtyp"`
		Oid      string   `json:"oid"`
		Rh       string   `json:"rh"`
		Sub      string   `json:"sub"`
		Tid      string   `json:"tid"`
		Uti      string   `json:"uti"`
		Ver      string   `json:"ver"`
		XmsTcdt  int      `json:"xms_tcdt"`
	}
)

var (
	envVarSplit        = regexp.MustCompile(`([\s,]+)`)
	subscriptionRegexp = regexp.MustCompile(`^(?i)/subscriptions/([^/]+)/?.*$`)
	providerRegexp     = regexp.MustCompile(`^(?i)/subscriptions/[^/]+(/resourcegroups/[^/]+)?/providers/([^/]+)/.*$`)

	tracingApiRequestEnabled      bool
	tracingLabelsApiEndpoint      bool
	tracingLabelsRoutingRegion    bool
	tracingLabelsSubscriptionID   bool
	tracingLabelsTenantID         bool
	tracingLabelsResourceProvider bool
	tracingLabelsMethod           bool
	tracingLabelsStatusCode       bool
	tracingApiRatelimitEnabled    bool
	tracingApiRatelimitAutoreset  bool
	tracingBuckets                = []float64{1, 5, 15, 30, 90}

	prometheusAzureApiRequest   *prometheus.HistogramVec
	prometheusAzureApiRatelimit *prometheus.GaugeVec
)

func TracingIsEnabled() bool {
	return tracingApiRatelimitEnabled || tracingApiRequestEnabled
}

func init() {
	// azureApiRequest settings
	tracingLabelsApiEndpoint = checkIfEnvVarContains(EnvVarApiRequestLables, "apiEndpoint", true)
	tracingLabelsRoutingRegion = checkIfEnvVarContains(EnvVarApiRequestLables, "routingRegion", false)
	tracingLabelsSubscriptionID = checkIfEnvVarContains(EnvVarApiRequestLables, "subscriptionID", true)
	tracingLabelsTenantID = checkIfEnvVarContains(EnvVarApiRequestLables, "tenantID", true)
	tracingLabelsResourceProvider = checkIfEnvVarContains(EnvVarApiRequestLables, "resourceProvider", true)
	tracingLabelsMethod = checkIfEnvVarContains(EnvVarApiRequestLables, "method", true)
	tracingLabelsStatusCode = checkIfEnvVarContains(EnvVarApiRequestLables, "statusCode", true)

	if envVal := os.Getenv(EnvVarApiRequestBuckets); envVal != "" {
		tracingBuckets = []float64{}
		for _, bucketString := range envVarSplit.Split(envVal, -1) {
			bucketString = strings.TrimSpace(bucketString)
			if val, err := strconv.ParseFloat(bucketString, 64); err == nil {
				tracingBuckets = append(
					tracingBuckets,
					val,
				)
			} else {
				panic(fmt.Sprintf("unable to parse env var %v=\"%v\": %v", EnvVarApiRequestBuckets, os.Getenv(EnvVarApiRequestBuckets), err))
			}
		}
	}

	// azureApiRatelimit
	tracingApiRequestEnabled = checkIfEnvVarIsEnabled(EnvVarApiRequestEnabled, true)
	tracingApiRatelimitEnabled = checkIfEnvVarIsEnabled(EnvVarApiRatelimitEnabled, true)
	tracingApiRatelimitAutoreset = checkIfEnvVarIsEnabled(EnvVarApiRatelimitAutoreset, true)

	labels := []string{}

	if tracingLabelsApiEndpoint {
		labels = append(labels, "apiEndpoint")
	}

	if tracingLabelsRoutingRegion {
		labels = append(labels, "routingRegion")
	}

	if tracingLabelsSubscriptionID {
		labels = append(labels, "subscriptionID")
	}

	if tracingLabelsTenantID {
		labels = append(labels, "tenantID")
	}

	if tracingLabelsResourceProvider {
		labels = append(labels, "resourceProvider")
	}

	if tracingLabelsMethod {
		labels = append(labels, "method")
	}

	if tracingLabelsStatusCode {
		labels = append(labels, "statusCode")
	}

	prometheusAzureApiRequest = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "azurerm_api_request",
			Help:    "AzureRM API requests",
			Buckets: tracingBuckets,
		},
		labels,
	)
	prometheus.MustRegister(prometheusAzureApiRequest)

	if tracingApiRatelimitEnabled {
		prometheusAzureApiRatelimit = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "azurerm_api_ratelimit",
				Help: "AzureRM API ratelimit",
			},
			[]string{
				"apiEndpoint",
				"subscriptionID",
				"tenantID",
				"scope",
				"type",
			},
		)
		prometheus.MustRegister(prometheusAzureApiRatelimit)
	}
}

func RegisterAzureMetricAutoClean(handler http.Handler) http.Handler {
	if prometheusAzureApiRatelimit == nil || !tracingApiRatelimitAutoreset {
		// metric or autoreset disabled, nothing to do here
		return handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		prometheusAzureApiRatelimit.Reset()
	})
}
