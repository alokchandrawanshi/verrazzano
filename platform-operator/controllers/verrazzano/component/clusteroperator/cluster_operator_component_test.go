// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package clusteroperator

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	vzconst "github.com/verrazzano/verrazzano/pkg/constants"
	"github.com/verrazzano/verrazzano/pkg/rancherutil"
	"github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1beta1"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/rancher"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"github.com/verrazzano/verrazzano/platform-operator/mocks"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const profilesRelativePath = "../../../../manifests/profiles"

// TestAppOperatorPostUpgradeNoDeleteClusterRoleBinding tests the PostUpgrade function
// GIVEN a call to PostUpgrade
// WHEN a VMC exists but no associated ClusterRoleBinding
// THEN no delete of a ClusterRoleBinding
func TestClusterOperatorEnabled(t *testing.T) {
	trueVal := true
	falseVal := false
	crA1 := &v1alpha1.Verrazzano{}
	crB1 := &v1beta1.Verrazzano{}

	crA1NilComp := crA1.DeepCopy()
	crA1NilComp.Spec.Components.ClusterOperator = nil
	crA1NilEnabled := crA1.DeepCopy()
	crA1NilEnabled.Spec.Components.ClusterOperator = &v1alpha1.ClusterOperatorComponent{Enabled: nil}
	crA1Enabled := crA1.DeepCopy()
	crA1Enabled.Spec.Components.ClusterOperator = &v1alpha1.ClusterOperatorComponent{Enabled: &trueVal}
	crA1Disabled := crA1.DeepCopy()
	crA1Disabled.Spec.Components.ClusterOperator = &v1alpha1.ClusterOperatorComponent{Enabled: &falseVal}

	crB1NilComp := crB1.DeepCopy()
	crB1NilComp.Spec.Components.ClusterOperator = nil
	crB1NilEnabled := crB1.DeepCopy()
	crB1NilEnabled.Spec.Components.ClusterOperator = &v1beta1.ClusterOperatorComponent{Enabled: nil}
	crB1Enabled := crB1.DeepCopy()
	crB1Enabled.Spec.Components.ClusterOperator = &v1beta1.ClusterOperatorComponent{Enabled: &trueVal}
	crB1Disabled := crB1.DeepCopy()
	crB1Disabled.Spec.Components.ClusterOperator = &v1beta1.ClusterOperatorComponent{Enabled: &falseVal}

	tests := []struct {
		name         string
		verrazzanoA1 *v1alpha1.Verrazzano
		verrazzanoB1 *v1beta1.Verrazzano
		assertion    func(t assert.TestingT, value bool, msgAndArgs ...interface{}) bool
	}{
		{
			name:         "test v1alpha1 component nil",
			verrazzanoA1: crA1NilComp,
			assertion:    assert.True,
		},
		{
			name:         "test v1alpha1 enabled nil",
			verrazzanoA1: crA1NilEnabled,
			assertion:    assert.True,
		},
		{
			name:         "test v1alpha1 enabled",
			verrazzanoA1: crA1Enabled,
			assertion:    assert.True,
		},
		{
			name:         "test v1alpha1 disabled",
			verrazzanoA1: crA1Disabled,
			assertion:    assert.False,
		},
		{
			name:         "test v1beta1 component nil",
			verrazzanoB1: crB1NilComp,
			assertion:    assert.True,
		},
		{
			name:         "test v1beta1 enabled nil",
			verrazzanoB1: crB1NilEnabled,
			assertion:    assert.True,
		},
		{
			name:         "test v1beta1 enabled",
			verrazzanoB1: crB1Enabled,
			assertion:    assert.True,
		},
		{
			name:         "test v1beta1 disabled",
			verrazzanoB1: crB1Disabled,
			assertion:    assert.False,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// clear any cached user auth tokens when the test completes
			defer rancherutil.DeleteStoredTokens()

			if tt.verrazzanoA1 != nil {
				tt.assertion(t, NewComponent().IsEnabled(spi.NewFakeContext(nil, tt.verrazzanoA1, tt.verrazzanoB1, false, profilesRelativePath).EffectiveCR()))
			}
			if tt.verrazzanoB1 != nil {
				tt.assertion(t, NewComponent().IsEnabled(spi.NewFakeContext(nil, tt.verrazzanoA1, tt.verrazzanoB1, false, profilesRelativePath).EffectiveCRV1Beta1()))
			}
		})
	}
}

// Test isReady when it's called with component context
func TestIsReady(t *testing.T) {
	c := fake.NewClientBuilder().Build()
	ctx := spi.NewFakeContext(c, &v1alpha1.Verrazzano{}, nil, true)
	assert.False(t, NewComponent().IsReady(ctx))
}

// Test isReady when it's called with component context when dry run false
func TestIsReadyFalse(t *testing.T) {
	c := fake.NewClientBuilder().Build()
	ctx := spi.NewFakeContext(c, &v1alpha1.Verrazzano{}, nil, false)
	assert.False(t, NewComponent().IsReady(ctx))
}

// TestPreInstall that the verrazzano-cluster-registrar ClusterRole has the VPO as its Helm release
// name, if it exists, after PreInstall is called.
func TestPreInstall(t *testing.T) {
	clustOpComp := clusterOperatorComponent{}
	runPreInstallUpgradeTests(t, clustOpComp.PreInstall)
}

// TestPreUpgrade that the verrazzano-cluster-registrar ClusterRole has the VPO as its Helm release
// name, if it exists, after PreUpgrade is called.
func TestPreUpgrade(t *testing.T) {
	clustOpComp := clusterOperatorComponent{}
	runPreInstallUpgradeTests(t, clustOpComp.PreUpgrade)
}

func runPreInstallUpgradeTests(t *testing.T, preInstallUpgradeFunc func(ctx spi.ComponentContext) error) {
	preInstallUpgradeTestCases := []struct {
		name              string
		clusterRoleExists bool
		existingHelmRel   string
		existingHelmNS    string
		expectErr         bool
	}{
		{"no ClusterRole exists", false, "", "", false},
		{"ClusterRole exists with no Helm annotations", true, "", "", false},
		{"ClusterRole exists with cluster operator Helm release", true, ComponentName, ComponentNamespace, false},
		{"ClusterRole exists with some other Helm release", true, "somerelease", "somenamespace", false},
		{"ClusterRole exists with VPO Helm release", true, constants.VerrazzanoPlatformOperatorHelmName, constants.VerrazzanoInstallNamespace, false},
	}
	for _, tt := range preInstallUpgradeTestCases {
		t.Run(tt.name, func(t *testing.T) {
			objs := []client.Object{}
			if tt.clusterRoleExists {
				annotations := map[string]string{}
				if tt.existingHelmRel != "" {
					annotations[helmReleaseNameAnnotation] = tt.existingHelmRel
				}
				if tt.existingHelmNS != "" {
					annotations[helmReleaseNamespaceAnnotation] = tt.existingHelmNS
				}
				clusterRole := rbacv1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{
						Name:        vzconst.VerrazzanoClusterRancherName,
						Annotations: annotations,
					},
				}
				objs = append(objs, &clusterRole)
			}
			c := fake.NewClientBuilder().WithObjects(objs...).Build()
			ctx := spi.NewFakeContext(c, &v1alpha1.Verrazzano{}, nil, true)
			err := preInstallUpgradeFunc(ctx)
			assert.NoError(t, err)
			clusterRoleNsn := types.NamespacedName{Name: vzconst.VerrazzanoClusterRancherName}
			updatedClusterRole := rbacv1.ClusterRole{}
			err = c.Get(context.TODO(), clusterRoleNsn, &updatedClusterRole)
			if tt.clusterRoleExists {
				assert.NoError(t, err)
				assert.Equal(t, constants.VerrazzanoPlatformOperatorHelmName, updatedClusterRole.Annotations[helmReleaseNameAnnotation])
				assert.Equal(t, constants.VerrazzanoInstallNamespace, updatedClusterRole.Annotations[helmReleaseNamespaceAnnotation])
			} else {
				assert.True(t, errors.IsNotFound(err))
			}
		})
	}
}

// TestPostInstall that the RoleTemplate gets created
func TestPostInstall(t *testing.T) {
	// clear any cached user auth tokens when the test completes
	defer rancherutil.DeleteStoredTokens()

	clustOpComp := clusterOperatorComponent{}

	cli := createClusterUserTestObjects().WithObjects(
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: vzconst.VerrazzanoClusterRancherName,
			},
		},
	).Build()

	mocker := gomock.NewController(t)
	httpMock := createClusterUserExists(mocks.NewMockRequestSender(mocker), http.StatusOK)

	savedRancherHTTPClient := rancherutil.RancherHTTPClient
	defer func() {
		rancherutil.RancherHTTPClient = savedRancherHTTPClient
	}()
	rancherutil.RancherHTTPClient = httpMock

	savedRetry := rancherutil.DefaultRetry
	defer func() {
		rancherutil.DefaultRetry = savedRetry
	}()
	rancherutil.DefaultRetry = wait.Backoff{
		Steps:    1,
		Duration: 1 * time.Millisecond,
		Factor:   1.0,
		Jitter:   0.1,
	}

	err := clustOpComp.PostInstall(spi.NewFakeContext(cli, &v1alpha1.Verrazzano{}, nil, false))
	assert.NoError(t, err)

	// Ensure the resource exists after postInstallUpgrade
	resource := unstructured.Unstructured{}
	resource.SetGroupVersionKind(rancher.GVKRoleTemplate)
	err = cli.Get(context.TODO(), types.NamespacedName{Name: vzconst.VerrazzanoClusterRancherName}, &resource)
	assert.NoError(t, err)
}

// TestPostUpgrade that the RoleTemplate gets created
func TestPostUpgrade(t *testing.T) {
	// clear any cached user auth tokens when the test completes
	defer rancherutil.DeleteStoredTokens()

	clustOpComp := clusterOperatorComponent{}

	cli := createClusterUserTestObjects().WithObjects(
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: vzconst.VerrazzanoClusterRancherName,
			},
		},
	).Build()

	mocker := gomock.NewController(t)
	httpMock := createClusterUserExists(mocks.NewMockRequestSender(mocker), http.StatusOK)

	savedRancherHTTPClient := rancherutil.RancherHTTPClient
	defer func() {
		rancherutil.RancherHTTPClient = savedRancherHTTPClient
	}()
	rancherutil.RancherHTTPClient = httpMock

	savedRetry := rancherutil.DefaultRetry
	defer func() {
		rancherutil.DefaultRetry = savedRetry
	}()
	rancherutil.DefaultRetry = wait.Backoff{
		Steps:    1,
		Duration: 1 * time.Millisecond,
		Factor:   1.0,
		Jitter:   0.1,
	}

	err := clustOpComp.PostUpgrade(spi.NewFakeContext(cli, &v1alpha1.Verrazzano{}, nil, false))
	assert.NoError(t, err)

	// Ensure the resource exists after postInstallUpgrade
	resource := unstructured.Unstructured{}
	resource.SetGroupVersionKind(rancher.GVKRoleTemplate)
	err = cli.Get(context.TODO(), types.NamespacedName{Name: vzconst.VerrazzanoClusterRancherName}, &resource)
	assert.NoError(t, err)
}
