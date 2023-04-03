// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	vzapi "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/common"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	admv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	deployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: common.CattleSystem,
			Name:      common.RancherName,
		},
	}
	daemonSet = &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: common.CattleSystem,
			Name:      common.RancherName,
		},
	}
	mutatingWebhookConfiguration = &admv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("rancher.%s", cattleNameFilter),
		},
	}
	validatingWebhookConfiguration = &admv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("rancher.%s", cattleNameFilter),
		},
	}
	mutatingWebhookConfiguration2 = &admv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("test-%s", webhookMonitorFilter),
		},
	}
	validatingWebhookConfiguration2 = &admv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("test-%s", webhookMonitorFilter),
		},
	}
	clusterRoleBinding1 = &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "clusterRoleBinding1",
			Labels: map[string]string{"cattle.io/creator": "norman"},
		},
	}
	clusterRole1 = &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "clusterRole1",
			Labels: map[string]string{"cattle.io/creator": "norman"},
		},
	}
	clusterRoleBinding2 = &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "cattle-clusterRoleBinding2",
			Labels: map[string]string{"clusterRoleBinding2": "true"},
		},
	}
	clusterRole2 = &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "cattle-clusterRole2",
			Labels: map[string]string{"clusterRole2": "true"},
		},
	}
	clusterRoleBinding3 = &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "rancher-clusterRoleBinding3",
			Labels: map[string]string{"clusterRoleBinding3": "true"},
		},
	}
	clusterRole3 = &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "rancher-clusterRole3",
			Labels: map[string]string{"clusterRole3": "true"},
		},
	}
	clusterRoleBinding4 = &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "fleet-clusterRoleBinding4",
			Labels: map[string]string{"clusterRoleBinding4": "true"},
		},
	}
	clusterRole4 = &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "fleet-clusterRole4",
			Labels: map[string]string{"clusterRole4": "true"},
		},
	}
)

// Test_cleanupPreventRecreate - test the cleanupPreventRecreate function
func Test_cleanupPreventRecreate(t *testing.T) {
	// Create fake client and context
	client := fake.NewClientBuilder().WithScheme(getSchemeForCleanup()).Build()
	ctx := spi.NewFakeContext(client, &vzapi.Verrazzano{}, nil, false)

	// Create a fake dynamic client
	fakeDynamicClient := dynfake.NewSimpleDynamicClient(getSchemeForCleanup(), newClusterCleanupRepoResources()...)

	// Override the dynamic client for unit testing and reset it when done
	prevGetDynamicClientFunc := getDynamicClientForCleanupFunc
	getDynamicClientForCleanupFunc = func() (dynamic.Interface, error) { return fakeDynamicClient, nil }
	defer func() {
		getDynamicClientForCleanupFunc = prevGetDynamicClientFunc
	}()

	// Verify expected resources exist prior to the cleanup
	verifyResources(t, ctx, fakeDynamicClient, false)

	// Call the function being tested
	cleanupRancher(ctx)

	// Verify expected resources do not exist after  the cleanup
	verifyResources(t, ctx, fakeDynamicClient, true)
}

// verifyResources - verify expected counts of resources before and after the rancher cleanup
func verifyResources(t *testing.T, ctx spi.ComponentContext, fakeDynamicClient dynamic.Interface, cleanupDone bool) {
	var expectedLen = 1
	if cleanupDone {
		expectedLen = 0
	}

	list, err := listResourceByNamespace(ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, ComponentNamespace, "")
	assert.NoError(t, err)
	assert.Equal(t, expectedLen, len(list.Items))

	list, err = listResourceByNamespace(ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}, ComponentNamespace, "")
	assert.NoError(t, err)
	assert.Equal(t, expectedLen, len(list.Items))

	list, err = listResource(ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: "v1", Resource: "mutatingwebhookconfigurations"}, "")
	assert.NoError(t, err)
	for _, item := range list.Items {
		if cleanupDone {
			assert.NotContains(t, item.GetName(), cattleNameFilter)
			assert.NotContains(t, item.GetName(), webhookMonitorFilter)
		} else {
			assert.True(t, strings.Contains(item.GetName(), cattleNameFilter) || strings.Contains(item.GetName(), webhookMonitorFilter))
		}
	}

	list, err = listResource(ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: "v1", Resource: "validatingwebhookconfigurations"}, "")
	assert.NoError(t, err)
	for _, item := range list.Items {
		if cleanupDone {
			assert.NotContains(t, item.GetName(), cattleNameFilter)
			assert.NotContains(t, item.GetName(), webhookMonitorFilter)
		} else {
			assert.True(t, strings.Contains(item.GetName(), cattleNameFilter) || strings.Contains(item.GetName(), webhookMonitorFilter))
		}
	}

	list, err = listResource(ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, normanSelector)
	assert.NoError(t, err)
	assert.Equal(t, expectedLen, len(list.Items))

	list, err = listResource(ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, normanSelector)
	assert.NoError(t, err)
	assert.Equal(t, expectedLen, len(list.Items))

	list, err = listResource(ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, "clusterRoleBinding2=true")
	assert.NoError(t, err)
	assert.Equal(t, expectedLen, len(list.Items))

	list, err = listResource(ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, "clusterRole2=true")
	assert.NoError(t, err)
	assert.Equal(t, expectedLen, len(list.Items))

	list, err = listResource(ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, "clusterRoleBinding3=true")
	assert.NoError(t, err)
	assert.Equal(t, expectedLen, len(list.Items))

	list, err = listResource(ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, "clusterRole3=true")
	assert.NoError(t, err)
	assert.Equal(t, expectedLen, len(list.Items))

	list, err = listResource(ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, "clusterRoleBinding4=true")
	assert.NoError(t, err)
	assert.Equal(t, expectedLen, len(list.Items))

	list, err = listResource(ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, "clusterRole4=true")
	assert.NoError(t, err)
	assert.Equal(t, expectedLen, len(list.Items))

}

func getSchemeForCleanup() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = admv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	return scheme
}

func newClusterCleanupRepoResources() []runtime.Object {
	return []runtime.Object{deployment, daemonSet, mutatingWebhookConfiguration, validatingWebhookConfiguration,
		mutatingWebhookConfiguration2, validatingWebhookConfiguration2, clusterRoleBinding1, clusterRole1,
		clusterRole2, clusterRoleBinding2, clusterRole3, clusterRoleBinding3, clusterRole4, clusterRoleBinding4}
}
