// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package system

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg"
	dump "github.com/verrazzano/verrazzano/tests/e2e/pkg/test/clusterdump"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/test/framework"
)

const (
	shortPollingInterval        = 10 * time.Second
	shortWaitTimeout            = 5 * time.Minute
	longPollingInterval         = 30 * time.Second
	longWaitTimeout             = 10 * time.Minute
	jaegerOperatorSampleMetric  = "jaeger_operator_instances_managed"
	jaegerCollectorSampleMetric = "jaeger_collector_queue_capacity"
)

var start time.Time

var t = framework.NewTestFramework("jaeger_mc_system_test")

var (
	adminKubeConfigPath = os.Getenv("ADMIN_KUBECONFIG")
	clusterName         = os.Getenv("CLUSTER_NAME")
	clusterNameLabel    string
	metricsTest         pkg.MetricsTest
	failed              = false
)

var beforeSuite = t.BeforeSuiteFunc(func() {
	// Allow 3hr allowance for the traces.
	start = time.Now().Add(-3 * time.Hour)
	if adminKubeConfigPath == "" {
		AbortSuite("Required env variable ADMIN_KUBECONFIG not set.")
	}

	var err error
	clusterNameLabel, err = pkg.GetClusterNameMetricLabel(adminKubeConfigPath)
	if err != nil {
		AbortSuite(fmt.Sprintf("Failed to get cluster label for metric collection: %v", err))
	}

	m := make(map[string]string)
	m[clusterNameLabel] = getClusterName()
	metricsTest, err = pkg.NewMetricsTest(adminKubeConfigPath, m)
	if err != nil {
		AbortSuite(fmt.Sprintf("Failed to create the Metrics test object: %v", err))
	}

})

var _ = BeforeSuite(beforeSuite)

var _ = t.AfterEach(func() {
	failed = failed || CurrentSpecReport().Failed()
})

var afterSuite = t.AfterSuiteFunc(func() {
	if failed {
		err := dump.ExecuteBugReport()
		if err != nil {
			pkg.Log(pkg.Error, err.Error())
		}
	}
})

var _ = AfterSuite(afterSuite)

var _ = t.Describe("Multi Cluster Jaeger Validation", Label("f:platform-lcm.install"), func() {

	// GIVEN a multicluster setup with an admin and a manged cluster,
	// WHEN Jaeger operator is enabled in the admin cluster and the managed cluster is registered to it,
	// THEN system traces can be queried from the Jaeger UI in the admin cluster
	t.It("traces from verrazzano system components of managed cluster should be available when queried from Jaeger", func() {
		validatorFn := pkg.ValidateSystemTracesFuncInCluster(adminKubeConfigPath, start, getClusterName())
		Eventually(validatorFn).WithPolling(longPollingInterval).WithTimeout(longWaitTimeout).Should(BeTrue())
	})

	// GIVEN a multicluster setup with an admin and a manged cluster,
	// WHEN Jaeger operator is enabled in the admin cluster and the managed cluster is registered to it,
	// THEN we are able to query the metrics of Jaeger operator running in managed cluster
	//      from the prometheus service running admin cluster.
	t.It("metrics of jaeger operator running in managed cluster are available in prometheus of admin cluster", func() {
		Eventually(func() bool {
			return metricsTest.MetricsExist(jaegerOperatorSampleMetric, map[string]string{})
		}).WithPolling(shortPollingInterval).WithTimeout(shortWaitTimeout).Should(BeTrue())
	})

	// GIVEN a multicluster setup with an admin and a manged cluster,
	// WHEN Jaeger operator is enabled in the admin cluster and the managed cluster is registered to it,
	// THEN we are able to query the metrics of Jaeger collector running in managed cluster
	//      from the prometheus service running admin cluster.
	t.It("metrics of jaeger collector running in managed cluster are available in prometheus of admin cluster", func() {
		Eventually(func() bool {
			return metricsTest.MetricsExist(jaegerCollectorSampleMetric, map[string]string{})
		}).WithPolling(shortPollingInterval).WithTimeout(shortWaitTimeout).Should(BeTrue())
	})
})

func getClusterName() string {
	if clusterName == "admin" {
		return "local"
	}
	return clusterName
}
