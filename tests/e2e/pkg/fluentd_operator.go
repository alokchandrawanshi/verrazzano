// Copyright (c) 2021, 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package pkg

import (
	"context"
	"fmt"

	globalconst "github.com/verrazzano/verrazzano/pkg/constants"
	"github.com/verrazzano/verrazzano/pkg/k8sutil"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const verrazzanoNamespace string = "verrazzano-system"
const VmiESURL = "http://verrazzano-authproxy-opensearch:8775"
const VmiESLegacySecret = "verrazzano"               //nolint:gosec //#gosec G101
const VmiESInternalSecret = "verrazzano-es-internal" //nolint:gosec //#gosec G101

func GetFluentBitDaemonset() (*appv1.DaemonSet, error) {
	clientset, err := k8sutil.GetKubernetesClientset()
	if err != nil {
		return nil, err
	}
	ds, err := clientset.AppsV1().DaemonSets(verrazzanoNamespace).Get(context.TODO(), globalconst.FluentBitDaemonSetName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return ds, nil
}

// IsFluentBitOperatorEnabled returns false if the Fluent Operator component is not set, or the value of its Enabled field otherwise
func IsFluentBitOperatorEnabled(kubeconfigPath string) bool {
	vz, err := GetVerrazzanoInstallResourceInCluster(kubeconfigPath)
	if err != nil {
		Log(Error, fmt.Sprintf("Error Verrazzano Resource: %v", err))
		return false
	}
	if vz == nil || vz.Spec.Components.FluentOperator == nil || vz.Spec.Components.FluentOperator.Enabled == nil {
		return false
	}
	return *vz.Spec.Components.FluentOperator.Enabled
}
