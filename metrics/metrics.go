package metrics

import (
	"net/url"
	"strconv"

	"github.com/TwiN/gatus/v5/config/endpoint"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "gatus" // The prefix of the metrics

var (
	initializedMetrics bool // Whether the metrics have been initialized

	resultTotal                        *prometheus.CounterVec
	resultDurationSeconds              *prometheus.GaugeVec
	resultConnectedTotal               *prometheus.CounterVec
	resultCodeTotal                    *prometheus.CounterVec
	resultCertificateExpirationSeconds *prometheus.GaugeVec
)

func initializePrometheusMetrics() {
	resultTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "results_total",
		Help:      "Number of results per endpoint",
	}, []string{"key", "group", "name", "type", "success", "url"})
	resultDurationSeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "results_duration_seconds",
		Help:      "Duration of the request in seconds",
	}, []string{"key", "group", "name", "type", "url"})
	resultConnectedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "results_connected_total",
		Help:      "Total number of results in which a connection was successfully established",
	}, []string{"key", "group", "name", "type", "url"})
	resultCodeTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "results_code_total",
		Help:      "Total number of results by code",
	}, []string{"key", "group", "name", "type", "code", "url"})
	resultCertificateExpirationSeconds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "results_certificate_expiration_seconds",
		Help:      "Number of seconds until the certificate expires",
	}, []string{"key", "group", "name", "type", "url"})
}

// PublishMetricsForEndpoint publishes metrics for the given endpoint and its result.
// These metrics will be exposed at /metrics if the metrics are enabled
func PublishMetricsForEndpoint(ep *endpoint.Endpoint, result *endpoint.Result) {
	if !initializedMetrics {
		initializePrometheusMetrics()
		initializedMetrics = true
	}
	endpointType := ep.Type()

	// Remove any query parameters from the URL to prevent any secrets from leaking in metrics
	url, err := url.Parse(ep.URL)
	var baseURL string
	if err != nil {
		baseURL = "-"
	} else {
		url.RawQuery = ""
		baseURL = url.String()
	}

	resultTotal.WithLabelValues(ep.Key(), ep.Group, ep.Name, string(endpointType), strconv.FormatBool(result.Success), baseURL).Inc()
	resultDurationSeconds.WithLabelValues(ep.Key(), ep.Group, ep.Name, string(endpointType), baseURL).Set(result.Duration.Seconds())
	if result.Connected {
		switch endpointType {
		case endpoint.TypeDNS:
			resultConnectedTotal.WithLabelValues(ep.Key(), ep.Group, ep.Name, string(endpointType), ep.DNSConfig.QueryName).Inc()
		default:
			resultConnectedTotal.WithLabelValues(ep.Key(), ep.Group, ep.Name, string(endpointType), baseURL).Inc()
		}
	}
	if result.DNSRCode != "" {
		resultCodeTotal.WithLabelValues(ep.Key(), ep.Group, ep.Name, string(endpointType), result.DNSRCode, ep.DNSConfig.QueryName).Inc()
	}
	if result.HTTPStatus != 0 {
		resultCodeTotal.WithLabelValues(ep.Key(), ep.Group, ep.Name, string(endpointType), strconv.Itoa(result.HTTPStatus), baseURL).Inc()
	}
	if result.CertificateExpiration != 0 {
		resultCertificateExpirationSeconds.WithLabelValues(ep.Key(), ep.Group, ep.Name, string(endpointType), baseURL).Set(result.CertificateExpiration.Seconds())
	}
}
