// Copyright (C) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package fluentbit_operator

import (
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/update"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/verrazzano/verrazzano/pkg/constants"
	vzapi "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/test/framework"
)

var trueValue = true
var falseValue = false

type FluentBitOperatorEnabledModifier struct {
}

func (u FluentBitOperatorEnabledModifier) ModifyCR(cr *vzapi.Verrazzano) {
	cr.Spec.Components.FluentOperator = &vzapi.FluentOperatorComponent{
		Enabled: &trueValue,
	}
	cr.Spec.Components.FluentbitOpensearchOutput = &vzapi.FluentbitOpensearchOutputComponent{
		Enabled: &trueValue,
	}
	cr.Spec.Components.Fluentd = &vzapi.FluentdComponent{
		Enabled: &falseValue,
	}
}

const (
	systemNamespace             = "verrazzano-system"
	longWaitTimeout             = 20 * time.Minute
	longPollingInterval         = 20 * time.Second
	shortPollingInterval        = 10 * time.Second
	shortWaitTimeout            = 5 * time.Minute
	pollingInterval             = 10 * time.Second
	waitTimeout                 = 20 * time.Minute
	fluentBitComponentLabel     = "app.kubernetes.io/name"
	fluentBitOperatorLabelValue = "fluentbit-operator"
	fluentBitLabelValue         = "fluentbit"
)

var (
	t = framework.NewTestFramework("install")
)
var _ = t.AfterEach(func() {})

var _ = BeforeSuite(beforeSuite)

var beforeSuite = t.BeforeSuiteFunc(func() {

	m := FluentBitOperatorEnabledModifier{}
	update.UpdateCRWithRetries(m, pollingInterval, waitTimeout)

	// GIVEN a VZ custom resource in dev profile,
	// WHEN Jaeger operator is enabled,
	// THEN Jaeger operator and pods for jaeger-collector and jaeger-query components gets created.
	update.ValidatePods(fluentBitOperatorLabelValue, fluentBitComponentLabel, constants.VerrazzanoSystemNamespace, 1, false)
	update.ValidatePods(fluentBitLabelValue, fluentBitComponentLabel, constants.VerrazzanoSystemNamespace, 1, false)
})

//	var _ = t.Describe("Verify fluentbit and configure VZ", func() {
//		t.It("verify fluentbit pods are ready", func() {
//			// Check all pods with fluentbit prefix
//			Eventually(func() bool {
//				isReady, err := pkg.PodsRunning(systemNamespace, []string{clusterName})
//				if err != nil {
//					return false
//				}
//				return isReady
//			}, longWaitTimeout, longPollingInterval).Should(BeTrue(), "FluentBit failed to get to ready state")
//		})
//	})
var _ = t.Describe("Verify OpenSearch infra", func() {
	t.It("verrazzano-system index is present", func() {
		Eventually(func() bool {
			return pkg.LogIndexFound("verrazzano-system")
		}, shortWaitTimeout, shortPollingInterval).Should(BeTrue())
	})

	pkg.Concurrently(
		func() {
			// GIVEN an application with logging enabled
			// WHEN the log records are retrieved from the Opensearch index for hello-helidon-container
			// THEN verify that at least one recent log record is found
			t.It("Verify recent Opensearch log record exists", func() {
				Eventually(func() bool {
					return pkg.LogRecordFound("verrazzano-system", time.Now().Add(-24*time.Minute), map[string]string{
						"kubernetes.container_name": "verrazzano-authproxy"})
				}, longWaitTimeout, longPollingInterval).Should(BeTrue(), "Expected to find a recent log record for container hello-helidon-container")
			})
		},
		func() {
			// GIVEN an application with logging enabled
			// WHEN the log records are retrieved from the Openearch index for other-container
			// THEN verify that at least one recent log record is found
			t.It("Verify recent Opensearch log record of other-container exists", func() {
				Eventually(func() bool {
					return pkg.LogRecordFound("verrazzano-system", time.Now().Add(-24*time.Minute), map[string]string{
						"kubernetes.container_name": "verrazzano-authproxy"})
				}, longWaitTimeout, longPollingInterval).Should(BeTrue(), "Expected to find a recent log record for other-container")
			})
		},
	)
})
