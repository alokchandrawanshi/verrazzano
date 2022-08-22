// Copyright (c) 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package mysqloperator

import (
	"fmt"

	vzapi "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"github.com/verrazzano/verrazzano/platform-operator/internal/k8s/status"
	"k8s.io/apimachinery/pkg/types"
)

// GetOverrides gets the install overrides
func GetOverrides(effectiveCR *vzapi.Verrazzano) []vzapi.Overrides {
	if effectiveCR.Spec.Components.MySQLOperator != nil {
		return effectiveCR.Spec.Components.MySQLOperator.ValueOverrides
	}
	return []vzapi.Overrides{}
}

// isReady - component specific checks for being ready
func isReady(ctx spi.ComponentContext) bool {
	return status.DeploymentsAreReady(ctx.Log(), ctx.Client(), getDeploymentList(), 1, getPrefix(ctx))
}

// isInstalled checks that the deployment exists
func isInstalled(ctx spi.ComponentContext) bool {
	return status.DoDeploymentsExist(ctx.Log(), ctx.Client(), getDeploymentList(), 1, getPrefix(ctx))
}

func getPrefix(ctx spi.ComponentContext) string {
	return fmt.Sprintf("Component %s", ctx.GetComponent())
}

func getDeploymentList() []types.NamespacedName {
	return []types.NamespacedName{
		{
			Name:      ComponentName,
			Namespace: ComponentNamespace,
		},
	}
}
