// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package fluentoperator

import (
	"fmt"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/verrazzano/verrazzano/pkg/bom"
	"github.com/verrazzano/verrazzano/pkg/k8s/ready"
	"github.com/verrazzano/verrazzano/pkg/k8sutil"
	"github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1beta1"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"github.com/verrazzano/verrazzano/platform-operator/internal/config"
)

const (
	fluentbitDaemonset         = "fluent-bit"
	fluentOperatorFullImageKey = "operator.container.repository"
	fluentBitFullImageKey      = "fluentbit.image.repository"
	fluentOperatorInitImageKey = "operator.initcontainer.repository"
	fluentOperatorImageTag     = "operator.container.tag"
	fluentOperatorInitTag      = "operator.initcontainer.tag"
	fluentBitImageTag          = "fluentbit.image.tag"
	clusterOutputDirectory     = "fluent-operator"
	systemOutputFile           = "system-output.yaml"
	applicationOutputFile      = "application-output.yaml"
)

var (
	componentPrefix          = fmt.Sprintf("Component %s", ComponentName)
	fluentOperatorDeployment = types.NamespacedName{
		Name:      ComponentName,
		Namespace: ComponentNamespace,
	}
	fluentBitDaemonSet = types.NamespacedName{
		Name:      fluentbitDaemonset,
		Namespace: ComponentNamespace,
	}
)

// isFluentOperatorReady checks if the Fluent operator is ready
func isFluentOperatorReady(context spi.ComponentContext) bool {
	return ready.DeploymentsAreReady(context.Log(), context.Client(), []types.NamespacedName{fluentOperatorDeployment}, 1, componentPrefix) &&
		ready.DaemonSetsAreReady(context.Log(), context.Client(), []types.NamespacedName{fluentBitDaemonSet}, 1, componentPrefix)
}

// GetOverrides returns install overrides for a component
func GetOverrides(object runtime.Object) interface{} {
	if effectiveCR, ok := object.(*v1alpha1.Verrazzano); ok {
		if effectiveCR.Spec.Components.FluentOperator != nil {
			return effectiveCR.Spec.Components.FluentOperator.ValueOverrides
		}
		return []v1alpha1.Overrides{}
	}
	effectiveCR := object.(*v1beta1.Verrazzano)
	if effectiveCR.Spec.Components.FluentOperator != nil {
		return effectiveCR.Spec.Components.FluentOperator.ValueOverrides
	}
	return []v1beta1.Overrides{}
}

// appendOverrides appends the overrides for the component
func appendOverrides(ctx spi.ComponentContext, _ string, _ string, _ string, kvs []bom.KeyValue) ([]bom.KeyValue, error) {
	bomFile, err := bom.NewBom(config.GetDefaultBOMFilePath())
	if err != nil {
		return kvs, ctx.Log().ErrorNewErr("Failed to get the BOM file for the fluent-operator image overrides: ", err)
	}
	images, err := bomFile.BuildImageOverrides("fluent-operator")
	if err != nil {
		return kvs, err
	}
	for _, image := range images {
		switch image.Key {
		case fluentOperatorFullImageKey:
			kvs = append(kvs, bom.KeyValue{Key: fluentOperatorFullImageKey, Value: image.Value})
		case fluentOperatorInitImageKey:
			kvs = append(kvs, bom.KeyValue{Key: fluentOperatorInitImageKey, Value: image.Value})
		case fluentBitFullImageKey:
			kvs = append(kvs, bom.KeyValue{Key: fluentBitFullImageKey, Value: image.Value})
		case fluentOperatorImageTag:
			kvs = append(kvs, bom.KeyValue{Key: fluentOperatorImageTag, Value: image.Value})
		case fluentOperatorInitTag:
			kvs = append(kvs, bom.KeyValue{Key: fluentOperatorInitTag, Value: image.Value})
		case fluentBitImageTag:
			kvs = append(kvs, bom.KeyValue{Key: fluentBitImageTag, Value: image.Value})
		}
	}
	if len(kvs) < 1 {
		return kvs, ctx.Log().ErrorfNewErr("Failed to construct fluent-operator related images from BOM")
	}
	kvs = append(kvs, bom.KeyValue{Key: "image.pullSecrets.enabled", Value: "true"})
	return kvs, nil
}

func applyClusterOutput(compContext spi.ComponentContext) error {
	crdManifestDir := filepath.Join(config.GetThirdPartyManifestsDir(), clusterOutputDirectory)
	systemClusterOutput := filepath.Join(crdManifestDir, systemOutputFile)
	applicationClusterOutput := filepath.Join(crdManifestDir, applicationOutputFile)
	// Apply the ClusterOutput for System related logs
	if err := k8sutil.NewYAMLApplier(compContext.Client(), "").ApplyF(systemClusterOutput); err != nil {
		return compContext.Log().ErrorfNewErr("Failed applying ClusterOutput for System related logs: %v", err)

	}
	if err := k8sutil.NewYAMLApplier(compContext.Client(), "").ApplyF(applicationClusterOutput); err != nil {
		return compContext.Log().ErrorfNewErr("Failed applying ClusterOutput for Application related logs: %v", err)
	}
	return nil
}