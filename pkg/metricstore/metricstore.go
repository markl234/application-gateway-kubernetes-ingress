// -------------------------------------------------------------------------------------------
// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
// --------------------------------------------------------------------------------------------

package metricstore

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Azure/application-gateway-kubernetes-ingress/pkg/annotations"
	"github.com/Azure/application-gateway-kubernetes-ingress/pkg/environment"
	"github.com/Azure/application-gateway-kubernetes-ingress/pkg/version"
)

// PrometheusNamespace is the namespace for appgw ingress controller
var PrometheusNamespace = "appgw_ingress_controller"

// MetricStore is store maintaining all metrics
type MetricStore interface {
	Start()
	Stop()
	Handler() http.Handler
	SetUpdateLatencySec(time.Duration)
	IncArmAPIUpdateCallFailureCounter()
	IncArmAPIUpdateCallSuccessCounter()
	IncArmAPICallCounter()
	IncK8sAPIEventCounter()
}

// AGICMetricStore is store
type AGICMetricStore struct {
	constLabels                    prometheus.Labels
	updateLatency                  prometheus.Gauge
	k8sAPIEventCounter             prometheus.Counter
	armAPICallCounter              prometheus.Counter
	armAPIUpdateCallFailureCounter prometheus.Counter
	armAPIUpdateCallSuccessCounter prometheus.Counter

	registry *prometheus.Registry
}

// NewMetricStore returns a new metric store
func NewMetricStore(envVariable environment.EnvVariables) MetricStore {
	constLabels := prometheus.Labels{
		"controller_class":                annotations.ApplicationGatewayIngressClass,
		"controller_namespace":            envVariable.AGICPodNamespace,
		"controller_pod":                  envVariable.AGICPodName,
		"controller_appgw_subscription":   envVariable.SubscriptionID,
		"controller_appgw_resource_group": envVariable.ResourceGroupName,
		"controller_appgw_name":           envVariable.AppGwName,
		"controller_version":              fmt.Sprintf("%s/%s/%s", version.Version, version.GitCommit, version.BuildDate),
	}
	return &AGICMetricStore{
		constLabels: constLabels,
		updateLatency: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   PrometheusNamespace,
			ConstLabels: constLabels,
			Name:        "update_latency_seconds",
			Help:        "The time spent in updating Application Gateway",
		}),
		k8sAPIEventCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   PrometheusNamespace,
			ConstLabels: constLabels,
			Name:        "k8s_api_event_counter",
			Help:        "This counter represents the number of events received from k8s API Server",
		}),
		armAPICallCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   PrometheusNamespace,
			ConstLabels: constLabels,
			Name:        "arm_api_call_counter",
			Help:        "This counter represents the number of API calls to ARM",
		}),
		armAPIUpdateCallFailureCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   PrometheusNamespace,
			ConstLabels: constLabels,
			Name:        "arm_api_update_call_failure_counter",
			Help:        "This counter represents the number of update API calls that failed to update Application Gateway",
		}),
		armAPIUpdateCallSuccessCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   PrometheusNamespace,
			ConstLabels: constLabels,
			Name:        "arm_api_update_call_success_counter",
			Help:        "This counter represents the number of update API calls that successfully updated Application Gateway",
		}),
		registry: prometheus.NewRegistry(),
	}
}

// Start store
func (ms *AGICMetricStore) Start() {
	ms.registry.MustRegister(ms.updateLatency)
}

// Stop store
func (ms *AGICMetricStore) Stop() {
	ms.registry.Unregister(ms.updateLatency)
}

// SetUpdateLatencySec updates latency
func (ms *AGICMetricStore) SetUpdateLatencySec(duration time.Duration) {
	ms.updateLatency.Set(duration.Seconds())
}

// IncK8sAPIEventCounter increases the counter after recieving a k8s Event
func (ms *AGICMetricStore) IncK8sAPIEventCounter() {
	ms.k8sAPIEventCounter.Inc()
}

// IncArmAPIUpdateCallFailureCounter increases the counter for failure on ARM
func (ms *AGICMetricStore) IncArmAPIUpdateCallFailureCounter() {
	ms.armAPIUpdateCallFailureCounter.Inc()
	ms.armAPICallCounter.Inc()
}

// IncArmAPIUpdateCallSuccessCounter increases the counter for success on ARM
func (ms *AGICMetricStore) IncArmAPIUpdateCallSuccessCounter() {
	ms.armAPIUpdateCallSuccessCounter.Inc()
	ms.armAPICallCounter.Inc()
}

// IncArmAPICallCounter increases the counter for success on ARM
func (ms *AGICMetricStore) IncArmAPICallCounter() {
	ms.armAPICallCounter.Inc()
}

// Handler return the registry
func (ms *AGICMetricStore) Handler() http.Handler {
	return promhttp.InstrumentMetricHandler(
		ms.registry,
		promhttp.HandlerFor(ms.registry, promhttp.HandlerOpts{}),
	)
}