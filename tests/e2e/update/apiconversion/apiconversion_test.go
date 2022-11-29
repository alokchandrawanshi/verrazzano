// Copyright (c) 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package apiconversion

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/verrazzano/verrazzano/pkg/constants"
	"github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1beta1"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/test/framework"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/update"
)

const (
	ingressNGINXComponentLabelKey        = "app.kubernetes.io/component"
	ingressNGINXComponentBackendValue    = "default-backend"
	ingressNGINXComponentControllerValue = "controller"
	ingressNGINXNameLabelValue           = "ingress-nginx"
	ingressNGINXNameLabelKey             = "app.kubernetes.io/name"
	waitTimeout                          = 10 * time.Minute
	pollingInterval                      = 5 * time.Second
)

type IngressNGINXControllerReplicasModifierV1beta1 struct {
	replicas uint32
}

type IngressNGINXBackendReplicasModifierV1beta1 struct {
	replicas uint32
}

type IngressNGINXDefaultModifierV1beta1 struct {
}

func (u IngressNGINXDefaultModifierV1beta1) ModifyCRV1beta1(cr *v1beta1.Verrazzano) {
	cr.Spec.Components.IngressNGINX = &v1beta1.IngressNginxComponent{}
}

var t = framework.NewTestFramework("apiconversion")

var nodeCount uint32

var beforeSuite = t.BeforeSuiteFunc(func() {
	var err error
	nodeCount, err = pkg.GetNodeCount()
	if err != nil {
		Fail(err.Error())
	}
})

var _ = BeforeSuite(beforeSuite)

var afterSuite = t.AfterSuiteFunc(func() {
	m := IngressNGINXDefaultModifierV1beta1{}
	update.UpdateCRV1beta1WithRetries(m, pollingInterval, waitTimeout)
	expectedRunning := uint32(2)
	update.ValidatePods(ingressNGINXNameLabelValue, ingressNGINXNameLabelKey, constants.IngressNamespace, expectedRunning, false)

})

var _ = AfterSuite(afterSuite)

func (u IngressNGINXBackendReplicasModifierV1beta1) ModifyCRV1beta1(cr *v1beta1.Verrazzano) {
	if cr.Spec.Components.IngressNGINX == nil {
		cr.Spec.Components.IngressNGINX = &v1beta1.IngressNginxComponent{}
	}
	ingressNginxReplicaOverridesYaml := fmt.Sprintf(`defaultBackend:
              replicaCount: %v`, u.replicas)
	cr.Spec.Components.IngressNGINX.ValueOverrides = pkg.CreateOverridesOrDie(ingressNginxReplicaOverridesYaml)
}

func (u IngressNGINXControllerReplicasModifierV1beta1) ModifyCRV1beta1(cr *v1beta1.Verrazzano) {
	if cr.Spec.Components.IngressNGINX == nil {
		cr.Spec.Components.IngressNGINX = &v1beta1.IngressNginxComponent{}
	}
	ingressNginxReplicaOverridesYaml := fmt.Sprintf(`controller:
              replicaCount: %v`, u.replicas)
	cr.Spec.Components.IngressNGINX.ValueOverrides = pkg.CreateOverridesOrDie(ingressNginxReplicaOverridesYaml)
}

var _ = t.Describe("Update ingressNGINX", Label("f:platform-lcm.update"), func() {
	t.Describe("ingressNginx update backend replicas with v1beta1 client", Label("f:platform-lcm.ingressNginx-update-replicas"), func() {
		t.It("ingressNginx explicit replicas", func() {
			m := IngressNGINXBackendReplicasModifierV1beta1{replicas: nodeCount}
			update.UpdateCRV1beta1WithRetries(m, pollingInterval, waitTimeout)
			expectedRunning := nodeCount
			update.ValidatePods(ingressNGINXComponentBackendValue, ingressNGINXComponentLabelKey, constants.IngressNamespace, expectedRunning, false)

		})
	})

	t.Describe("ingressNginx update controller replicas with v1beta1 client", Label("f:platform-lcm.ingressNginx-update-replicas"), func() {
		t.It("ingressNginx explicit replicas", func() {
			m := IngressNGINXControllerReplicasModifierV1beta1{replicas: nodeCount}
			update.UpdateCRV1beta1WithRetries(m, pollingInterval, waitTimeout)
			expectedRunning := nodeCount
			update.ValidatePods(ingressNGINXComponentControllerValue, ingressNGINXComponentLabelKey, constants.IngressNamespace, expectedRunning, false)

		})
	})
})
