// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package grafanadb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/verrazzano/verrazzano/pkg/constants"
	"github.com/verrazzano/verrazzano/pkg/k8sutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/test/framework"
)

const (
	waitTimeout     = 3 * time.Minute
	pollingInterval = 10 * time.Second
	documentFile    = "testdata/upgrade/grafana/dashboard.json"
)

var clientset = k8sutil.GetKubernetesClientsetOrDie()
var testDashboard pkg.DashboardMetadata
var t = framework.NewTestFramework("grafana")

var beforeSuite = t.BeforeSuiteFunc(func() {
	kubeconfigPath, err := k8sutil.GetKubeConfigLocation()
	if err != nil {
		Fail(fmt.Sprintf(pkg.KubeConfigErrorFmt, err))
	}
	supported := pkg.IsGrafanaEnabled(kubeconfigPath)
	// Only run tests if Grafana component is enabled in Verrazzano CR
	if !supported {
		Skip("Grafana component is not enabled")
	}
	// Create the test Grafana dashboard
	file, err := pkg.FindTestDataFile(documentFile)
	if err != nil {
		pkg.Log(pkg.Error, fmt.Sprintf("failed to find test data file: %v", err))
		Fail(err.Error())
	}
	data, err := os.ReadFile(file)
	if err != nil {
		pkg.Log(pkg.Error, fmt.Sprintf("failed to read test data file: %v", err))
		Fail(err.Error())
	}
	Eventually(func() bool {
		resp, err := pkg.CreateGrafanaDashboard(string(data))
		if err != nil {
			pkg.Log(pkg.Error, fmt.Sprintf("Failed to create Grafana testDashboard: %v", err))
			return false
		}
		if resp.StatusCode != http.StatusOK {
			pkg.Log(pkg.Error, fmt.Sprintf("Failed to create Grafana testDashboard: status=%d: body=%s", resp.StatusCode, string(resp.Body)))
			return false
		}
		json.Unmarshal(resp.Body, &testDashboard)
		return true
	}).WithPolling(pollingInterval).WithTimeout(waitTimeout).Should(BeTrue(),
		"It should be possible to create a Grafana dashboard and persist it.")
})

var _ = BeforeSuite(beforeSuite)

var _ = t.Describe("Test Grafana Dashboard Persistence", Label("f:observability.logging.es"), func() {

	// GIVEN a running grafana instance,
	// WHEN a GET call is made  to Grafana with the UID of the newly created testDashboard,
	// THEN the testDashboard metadata of the corresponding testDashboard is returned.
	It("Get details of the test Grafana Dashboard", func() {
		pkg.TestGrafanaTestDashboard(testDashboard, pollingInterval, waitTimeout)
	})

	// GIVEN a running Grafana instance,
	// WHEN a search is done based on the dashboard title,
	// THEN the details of the dashboards matching the search query is returned.
	It("Search the test Grafana Dashboard using its title", func() {
		pkg.TestSearchGrafanaDashboard(pollingInterval, waitTimeout)
	})

	// GIVEN a running grafana instance,
	// WHEN a GET call is made  to Grafana with the system dashboard UID,
	// THEN the dashboard metadata of the corresponding testDashboard is returned.
	It("Get details of the system Grafana Dashboard", func() {
		pkg.TestSystemHealthGrafanaDashboard(pollingInterval, waitTimeout)
	})

	kubeconfigPath, err := k8sutil.GetKubeConfigLocation()
	if err != nil {
		Expect(err).To(BeNil(), fmt.Sprintf(pkg.KubeConfigErrorFmt, err))
	}

	// GIVEN a running grafana instance
	// WHEN a call is made to Grafana Dashboard with UID corresponding to OpenSearch Summary Dashboard
	// THEN the dashboard metadata of the corresponding dashboard is returned
	if ok, _ := pkg.IsVerrazzanoMinVersion("1.3.0", kubeconfigPath); ok {
		t.It("Get details of the OpenSearch Grafana Dashboard", func() {
			pkg.TestOpenSearchGrafanaDashBoard(pollingInterval, waitTimeout)
		})
	}

	// GIVEN a running grafana instance,
	// WHEN the pod is deleted,
	// THEN pod eventually comes back up.
	It("Delete and wait for the pod to come back up", func() {
		// delete grafana pods
		Eventually(func() error {
			pods, err := clientset.CoreV1().Pods(constants.VerrazzanoSystemNamespace).List(context.TODO(), metav1.ListOptions{
				LabelSelector: "app=system-grafana",
			})
			if err != nil {
				pkg.Log(pkg.Error, "Failed to find grafana pod")
				return err
			}

			for i := range pods.Items {
				pod := &pods.Items[i]
				if err := clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{}); err != nil {
					pkg.Log(pkg.Error, "Failed to delete grafana pod")
					return err
				}
			}

			return nil
		}).WithPolling(pollingInterval).WithTimeout(waitTimeout).ShouldNot(HaveOccurred())

		// wait for pods to come back up
		Eventually(func() (bool, error) {
			pods, err := clientset.CoreV1().Pods(constants.VerrazzanoSystemNamespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				pkg.Log(pkg.Info, "Failed to get grafana pod")
				return false, err
			}

			if pods == nil || len(pods.Items) == 0 {
				return false, nil
			}
			for _, pod := range pods.Items {
				if !IsPodReadyOrCompleted(pod) {
					return false, nil
				}
			}

			return true, nil
		}).WithPolling(pollingInterval).WithTimeout(waitTimeout).Should(BeTrue(), "Expected Grafana pods to restart")
	})

	// GIVEN a running Grafana instance,
	// WHEN a search is made for the dashboard using its title,
	// THEN the dashboard metadata is returned.
	It("Search the test Grafana Dashboard using its title", func() {
		pkg.TestSearchGrafanaDashboard(pollingInterval, waitTimeout)
	})

	// GIVEN a running grafana instance,
	// WHEN a GET call is made  to Grafana with the UID of the system dashboard,
	// THEN the dashboard metadata of the corresponding System dashboard is returned.
	It("Get details of the system Grafana dashboard", func() {
		pkg.TestSystemHealthGrafanaDashboard(pollingInterval, waitTimeout)
	})

	kubeconfigPath, err = k8sutil.GetKubeConfigLocation()
	if err != nil {
		Expect(err).To(BeNil(), fmt.Sprintf(pkg.KubeConfigErrorFmt, err))
	}

	// GIVEN a running grafana instance
	// WHEN a call is made to Grafana Dashboard with UID corresponding to OpenSearch Summary Dashboard
	// THEN the dashboard metadata of the corresponding dashboard is returned
	if ok, _ := pkg.IsVerrazzanoMinVersion("1.3.0", kubeconfigPath); ok {
		t.It("Get details of the OpenSearch Grafana Dashboard", func() {
			pkg.TestOpenSearchGrafanaDashBoard(pollingInterval, waitTimeout)
		})
	}
})

func IsPodReadyOrCompleted(pod corev1.Pod) bool {
	switch pod.Status.Phase {
	case corev1.PodSucceeded:
		return true
	case corev1.PodRunning:
		for _, c := range pod.Status.ContainerStatuses {
			if !c.Ready {
				return false
			}
		}
		return true
	default:
		return false
	}
}
