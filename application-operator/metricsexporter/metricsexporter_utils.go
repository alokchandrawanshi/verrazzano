// Copyright (c) 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package metricsexporter

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	vzlogInit "github.com/verrazzano/verrazzano/pkg/log"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
)

type metricName string

const (
	AppconfigReconcileCounter              metricName = "appconfig reconcile counter"
	AppconfigReconcileError                metricName = "appconfig reconcile error"
	AppconfigReconcileDuration             metricName = "appconfig reconcile duration"
	CohworkloadReconcileCounter            metricName = "coherence reconcile counter"
	CohworkloadReconcileError              metricName = "coherence reconcile error"
	CohworkloadReconcileDuration           metricName = "coherence reconcile duration"
	HelidonReconcileCounter                metricName = "helidon reconcile counter"
	HelidonReconcileError                  metricName = "helidon reconcile error"
	HelidonReconcileDuration               metricName = "helidon reconcile duration"
	IngresstraitReconcileCounter           metricName = "ingress reconcile counter"
	IngresstraitReconcileError             metricName = "ingress reconcile error"
	IngresstraitReconcileDuration          metricName = "ingress reconcile duration"
	AppconfigHandleCounter                 metricName = "appconfig handle counter"
	AppconfigHandleError                   metricName = "appconfig handle error"
	AppconfigHandleDuration                metricName = "appconfig handle duration"
	IstioHandleCounter                     metricName = "istio handle counter"
	IstioHandleError                       metricName = "istio handle error"
	IstioHandleDuration                    metricName = "istio handle duration"
	LabelerPodHandleCounter                metricName = "LabelerPod handle counter"
	LabelerPodHandleError                  metricName = "LabelerPod handle error"
	LabelerPodHandleDuration               metricName = "LabelerPod handle duration"
	BindingUpdaterHandleCounter            metricName = "BindingUpdater handle counter"
	BindingUpdaterHandleError              metricName = "BindingUpdater handle error"
	BindingUpdaterHandleDuration           metricName = "BindingUpdater handle duration"
	MultiClusterAppconfigPodHandleCounter  metricName = "MultiClusterAppconfig handle counter"
	MultiClusterAppconfigPodHandleError    metricName = "MultiClusterAppconfig handle error"
	MultiClusterAppconfigPodHandleDuration metricName = "MultiClusterAppconfig handle duration"
	MultiClusterCompHandleCounter          metricName = "MultiClusterComp handle counter"
	MultiClusterCompHandleError            metricName = "MultiClusterComp handle error"
	MultiClusterCompHandleDuration         metricName = "MultiClusterComp handle duration"
	MultiClusterConfigmapHandleCounter     metricName = "MultiClusterConfigmap handle counter"
	MultiClusterConfigmapHandleError       metricName = "MultiClusterConfigmap handle error"
	MultiClusterConfigmapHandleDuration    metricName = "MultiClusterConfigmap handle duration"
	MultiClusterSecretHandleCounter        metricName = "MultiClusterSecret handle counter"
	MultiClusterSecretHandleError          metricName = "MultiClusterSecret handle error"
	MultiClusterSecretHandleDuration       metricName = "MultiClusterSecret handle duration"
	VzProjHandleCounter                    metricName = "VzProj handle counter"
	VzProjHandleError                      metricName = "VzProj handle error"
	VzProjHandleDuration                   metricName = "VzProj handle duration"
)

func init() {
	RequiredInitialization()
	RegisterMetrics()
}

// RequiredInitialization initializes the metrics object, but does not register the metrics
func RequiredInitialization() {
	MetricsExp = metricsExporter{
		internalConfig: initConfiguration(),
		internalData: data{
			simpleCounterMetricMap: initCounterMetricMap(),
			durationMetricMap:      initDurationMetricMap(),
		},
	}
}

// RegisterMetrics begins the process of registering metrics
func RegisterMetrics() {
	InitializeAllMetricsArray()
	go registerMetricsHandlers(zap.S())
}

// InitializeAllMetricsArray initializes the allMetrics array
func InitializeAllMetricsArray() {
	// Loop through all metrics declarations in metric maps
	for _, value := range MetricsExp.internalData.simpleCounterMetricMap {
		MetricsExp.internalConfig.allMetrics = append(MetricsExp.internalConfig.allMetrics, value.metric)
	}
	for _, value := range MetricsExp.internalData.durationMetricMap {
		MetricsExp.internalConfig.allMetrics = append(MetricsExp.internalConfig.allMetrics, value.metric)
	}

}

// initCounterMetricMap initializes the simpleCounterMetricMap for the metricsExporter object
func initCounterMetricMap() map[metricName]*SimpleCounterMetric {
	return map[metricName]*SimpleCounterMetric{
		AppconfigReconcileCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_appconfig_successful_reconcile_total",
				Help: "Tracks how many times the appconfig reconcile process has been successful"}),
		},
		AppconfigReconcileError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_appconfig_error_reconcile_total",
				Help: "Tracks how many times the appconfig reconcile process has failed"}),
		},
		CohworkloadReconcileCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_cohworkload_successful_reconcile_total",
				Help: "Tracks how many times the cohworkload reconcile process has been successful"}),
		},
		CohworkloadReconcileError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_cohworkload_error_reconcile_total",
				Help: "Tracks how many times the cohworkload reconcile process has failed"}),
		},
		HelidonReconcileCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_helidonworkload_successful_reconcile_total",
				Help: "Tracks how many times the helidonworkload reconcile process has been successful"}),
		},
		HelidonReconcileError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_helidonworkload_error_reconcile_total",
				Help: "Tracks how many times the helidonworkload reconcile process has failed"}),
		},
		IngresstraitReconcileCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_ingresstrait_successful_reconcile_total",
				Help: "Tracks how many times the ingresstrait reconcile process has been successful"}),
		},
		IngresstraitReconcileError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_ingresstrait_error_reconcile_total",
				Help: "Tracks how many times the ingresstrait reconcile process has failed"}),
		},
		AppconfigHandleCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_appconfig_handle_total",
				Help: "Tracks how many times appconfig handle process has been successful"}),
		},
		AppconfigHandleError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_appconfig_error_handle_total",
				Help: "Tracks how many times appconfig handle process has failed"}),
		},
		IstioHandleCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_istio_handle_total",
				Help: "Tracks how many times istio handle process has been successful"}),
		},
		IstioHandleError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_istio_error_handle_total",
				Help: "Tracks how many times istio handle process has failed"}),
		},
		LabelerPodHandleCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_labelerPod_handle_total",
				Help: "Tracks how many times the labeler pod handle process has been successful"}),
		},
		LabelerPodHandleError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_labelerpod_error_handle_total",
				Help: "Tracks how many times the labeler pod handle process has failed"}),
		},
		BindingUpdaterHandleCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_bindingupdater_handle_total",
				Help: "Tracks how many times the binding updater handle process has been successful"}),
		},
		BindingUpdaterHandleError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_bindingupdater_error_handle_total",
				Help: "Tracks how many times the binding updater handle process has failed"}),
		},
		MultiClusterAppconfigPodHandleCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_multiclusterappconfig_handle_total",
				Help: "Tracks how many times the multicluster appconfig pod handle process has been successful"}),
		},
		MultiClusterAppconfigPodHandleError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_multiclusterappconfig_error_handle_total",
				Help: "Tracks how many times the multicluster appconfig pod handle process has failed"}),
		},
		MultiClusterCompHandleCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_multiclustercomp_handle_total",
				Help: "Tracks how many times the multicluster component handle process has been successful"}),
		},
		MultiClusterCompHandleError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_multiclustercomp_error_handle_total",
				Help: "Tracks how many times the multicluster component handle process has failed"}),
		},
		MultiClusterConfigmapHandleCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_multiclustercomp_handle_total",
				Help: "Tracks how many times the multicluster configmap handle process has been successful"}),
		},
		MultiClusterConfigmapHandleError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_multiclustercomp_error_handle_total",
				Help: "Tracks how many times the multicluster configmap handle process has failed"}),
		},
		MultiClusterSecretHandleCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_multiclustersecret_handle_total",
				Help: "Tracks how many times the multicluster secret handle process has been successful"}),
		},
		MultiClusterSecretHandleError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_multiclustersecret_error_handle_total",
				Help: "Tracks how many times the multicluster secret handle process has failed"}),
		},
		VzProjHandleCounter: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_vzproj_handle_total",
				Help: "Tracks how many times the vz project handle process has been successful"}),
		},
		VzProjHandleError: {
			metric: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "vz_application_operator_vzproj_error_handle_total",
				Help: "Tracks how many times the vz project handle process has failed"}),
		},
	}
}

// initDurationMetricMap initializes the DurationMetricMap for the metricsExporter object
func initDurationMetricMap() map[metricName]*DurationMetrics {
	return map[metricName]*DurationMetrics{
		AppconfigReconcileDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_appconfig_reconcile_duration",
				Help: "The duration in seconds of vao appconfig reconcile process",
			}),
		},
		CohworkloadReconcileDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_cohworkload_reconcile_duration",
				Help: "The duration in seconds of vao coherence workload reconcile process",
			}),
		},
		HelidonReconcileDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_helidon_reconcile_duration",
				Help: "The duration in seconds of vao helidon reconcile process",
			}),
		},
		IngresstraitReconcileDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_ingresstrait_reconcile_duration",
				Help: "The duration in seconds of vao ingresstrait reconcile process",
			}),
		},
		AppconfigHandleDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_appconfig_handle_duration",
				Help: "The duration in seconds of vao appconfig handle process",
			}),
		},
		IstioHandleDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_istio_handle_duration",
				Help: "The duration in seconds of vao istio handle process",
			}),
		},
		LabelerPodHandleDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_labelerpod_handle_duration",
				Help: "The duration in seconds of vao labeler pod handle process",
			}),
		},
		MultiClusterConfigmapHandleDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_multiclusterconfigmap_handle_duration",
				Help: "The duration in seconds of vao multicluster configmap handle process",
			}),
		},
		MultiClusterAppconfigPodHandleDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_multiclusterappconfig_handle_duration",
				Help: "The duration in seconds of vao multicluster appconfig process",
			}),
		},
		MultiClusterCompHandleDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_multiclustercomp_handle_duration",
				Help: "The duration in seconds of vao multicluster component handle process",
			}),
		},
		MultiClusterSecretHandleDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_multiclustersecret_handle_duration",
				Help: "The duration in seconds of vao multicluster secret handle process",
			}),
		},
		VzProjHandleDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_vzproj_handle_duration",
				Help: "The duration in seconds of vao vz project handle process",
			}),
		},
		BindingUpdaterHandleDuration: {
			metric: prometheus.NewSummary(prometheus.SummaryOpts{
				Name: "vz_application_operator_bindingupdater_handle_duration",
				Help: "The duration in seconds of vao binding updater handle process",
			}),
		},
	}
}

// registerMetricsHandlersHelper is a helper function that assists in registering metrics
func registerMetricsHandlersHelper() error {
	var errorObserved error
	for metric := range MetricsExp.internalConfig.failedMetrics {
		err := MetricsExp.internalConfig.registry.Register(metric)
		if err != nil {
			if errorObserved != nil {
				errorObserved = errors.Wrap(errorObserved, err.Error())
			} else {
				errorObserved = err
			}
		} else {
			// If a metric is registered, delete it from the failed metrics map so that it is not retried
			delete(MetricsExp.internalConfig.failedMetrics, metric)
		}
	}
	return errorObserved
}

// registerMetricsHandlers registers the metrics and provides error handling
func registerMetricsHandlers(log *zap.SugaredLogger) {
	// Get list of metrics to register initially
	initializeFailedMetricsArray()
	// Loop until there is no error in registering
	for err := registerMetricsHandlersHelper(); err != nil; err = registerMetricsHandlersHelper() {
		log.Infof("Failed to register metrics for VMI %v", err)
		time.Sleep(time.Second)
	}
}

// initializeFailedMetricsArray initializes the failedMetrics array
func initializeFailedMetricsArray() {
	for i, metric := range MetricsExp.internalConfig.allMetrics {
		MetricsExp.internalConfig.failedMetrics[metric] = i
	}
}

// StartMetricsServer starts the metric server to begin emitting metrics to Prometheus
func StartMetricsServer() error {
	vlog, err := vzlog.EnsureResourceLogger(&vzlog.ResourceConfig{
		Name:           "",
		Namespace:      "",
		ID:             "",
		Generation:     0,
		ControllerName: "metricsexporter",
	})
	if err != nil {
		return err
	}
	go wait.Until(func() {
		http.Handle("/metrics", promhttp.Handler())
		server := &http.Server{
			Addr:              ":9100",
			ReadHeaderTimeout: 3 * time.Second,
		}
		err := server.ListenAndServe()
		if err != nil {
			vlog.Oncef("Failed to start metrics server for VMI: %v", err)
		}
	}, time.Second*3, wait.NeverStop)
	return nil
}

// initConfiguration returns an empty struct of type configuration
func initConfiguration() configuration {
	return configuration{
		allMetrics:    []prometheus.Collector{},
		failedMetrics: map[prometheus.Collector]int{},
		registry:      prometheus.DefaultRegisterer,
	}
}

// GetSimpleCounterMetric returns a simpleCounterMetric from the simpleCounterMetricMap given a metricName
func GetSimpleCounterMetric(name metricName) (*SimpleCounterMetric, error) {
	counterMetric, ok := MetricsExp.internalData.simpleCounterMetricMap[name]
	if !ok {
		return nil, fmt.Errorf("%v not found in SimpleCounterMetricMap due to metricName being defined, but not being a key in the map", name)
	}
	return counterMetric, nil
}

// GetDurationMetric returns a durationMetric from the durationMetricMap given a metricName
func GetDurationMetric(name metricName) (*DurationMetrics, error) {
	durationMetric, ok := MetricsExp.internalData.durationMetricMap[name]
	if !ok {
		return nil, fmt.Errorf("%v not found in durationMetricMap due to metricName being defined, but not being a key in the map", name)
	}
	return durationMetric, nil
}
func ExposeControllerMetrics(controllerName string, successname metricName, errorname metricName, durationname metricName) (*SimpleCounterMetric, *SimpleCounterMetric, *DurationMetrics, *zap.SugaredLogger, error) {
	zapLogForMetrics := zap.S().With(vzlogInit.FieldController, controllerName)
	counterMetricObject, err := GetSimpleCounterMetric(successname)
	if err != nil {
		zapLogForMetrics.Error(err)
		return nil, nil, nil, nil, err
	}
	errorCounterMetricObject, err := GetSimpleCounterMetric(errorname)
	if err != nil {
		zapLogForMetrics.Error(err)
		return nil, nil, nil, nil, err
	}

	durationMetricObject, err := GetDurationMetric(durationname)
	if err != nil {
		zapLogForMetrics.Error(err)
		return nil, nil, nil, nil, err
	}
	return counterMetricObject, errorCounterMetricObject, durationMetricObject, zapLogForMetrics, nil
}
