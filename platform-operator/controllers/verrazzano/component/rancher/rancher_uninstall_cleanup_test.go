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
	policyv1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	emptyFinalizer                  []string
	namespace1                      = newNamespace("local", map[string]string{"namespace1": "true"})
	namespace2                      = newNamespace("cattle-system", map[string]string{"namespace2": "true"})
	namespace3                      = newNamespace("cattle-impersonation-system", map[string]string{"namespace3": "true"})
	namespace4                      = newNamespace("cattle-global-data", map[string]string{"namespace4": "true"})
	namespace5                      = newNamespace("cattle-global-nt", map[string]string{"namespace5": "true"})
	namespace6                      = newNamespace("cattle-resources-system", map[string]string{"namespace6": "true"})
	namespace7                      = newNamespace("cis-operator-system", map[string]string{"namespace7": "true"})
	namespace8                      = newNamespace("cattle-dashboards", map[string]string{"namespace8": "true"})
	namespace9                      = newNamespace("cattle-gatekeeper-system", map[string]string{"namespace9": "true"})
	namespace10                     = newNamespace("cattle-alerting", map[string]string{"namespace10": "true"})
	namespace11                     = newNamespace("cattle-logging", map[string]string{"namespace11": "true"})
	namespace12                     = newNamespace("cattle-pipeline", map[string]string{"namespace12": "true"})
	namespace13                     = newNamespace("cattle-prometheus", map[string]string{"namespace13": "true"})
	namespace14                     = newNamespace("rancher-operator-system", map[string]string{"namespace14": "true"})
	namespace15                     = newNamespace("cattle-monitoring-system", map[string]string{"namespace15": "true"})
	namespace16                     = newNamespace("cattle-logging-system", map[string]string{"namespace16": "true"})
	namespace17                     = newNamespace("cattle-fleet-clusters-system", map[string]string{"namespace17": "true"})
	namespace18                     = newNamespace("cattle-fleet-local-system", map[string]string{"namespace18": "true"})
	namespace19                     = newNamespace("cattle-fleet-system", map[string]string{"namespace19": "true"})
	namespace20                     = newNamespace("fleet-default", map[string]string{"namespace20": "true"})
	namespace21                     = newNamespace("fleet-local", map[string]string{"namespace21": "true"})
	namespace22                     = newNamespace("fleet-system", map[string]string{"namespace22": "true"})
	namespace23                     = newNamespace("cluster-fleet", map[string]string{"namespace23": "true"})
	namespace24                     = newNamespace("c-test", map[string]string{"namespace24": "true"})
	namespace25                     = newNamespace("p-test", map[string]string{"namespace25": "true"})
	namespace26                     = newNamespace("user-test", map[string]string{"namespace26": "true"})
	namespace27                     = newNamespace("u-test", map[string]string{"namespace27": "true"})
	deployment                      = newDeployment(common.CattleSystem, common.RancherName)
	daemonSet                       = newDaemonSet(common.CattleSystem, common.RancherName)
	mutatingWebhookConfiguration    = newMutatingWebhookConfiguration(fmt.Sprintf("rancher.%s", cattleNameFilter))
	validatingWebhookConfiguration  = newValidatingWebhookConfiguration(fmt.Sprintf("rancher.%s", cattleNameFilter))
	mutatingWebhookConfiguration2   = newMutatingWebhookConfiguration(fmt.Sprintf("test-%s", webhookMonitorFilter))
	validatingWebhookConfiguration2 = newValidatingWebhookConfiguration(fmt.Sprintf("test-%s", webhookMonitorFilter))
	clusterRoleBinding1             = newClusterRoleBinding("clusterRoleBinding1", map[string]string{"cattle.io/creator": "norman"}, emptyFinalizer)
	clusterRole1                    = newClusterRole("clusterRole1", map[string]string{"cattle.io/creator": "norman"}, emptyFinalizer)
	clusterRoleBinding2             = newClusterRoleBinding("cattle-clusterRoleBinding2", map[string]string{"clusterRoleBinding2": "true"}, emptyFinalizer)
	clusterRole2                    = newClusterRole("cattle-clusterRole2", map[string]string{"clusterRole2": "true"}, emptyFinalizer)
	clusterRoleBinding3             = newClusterRoleBinding("rancher-clusterRoleBinding3", map[string]string{"clusterRoleBinding3": "true"}, emptyFinalizer)
	clusterRole3                    = newClusterRole("rancher-clusterRole3", map[string]string{"clusterRole3": "true"}, emptyFinalizer)
	clusterRoleBinding4             = newClusterRoleBinding("fleet-clusterRoleBinding4", map[string]string{"clusterRoleBinding4": "true"}, emptyFinalizer)
	clusterRole4                    = newClusterRole("fleet-clusterRole4", map[string]string{"clusterRole4": "true"}, emptyFinalizer)
	clusterRoleBinding5             = newClusterRoleBinding("gitjob-clusterRoleBinding5", map[string]string{"clusterRoleBinding5": "true"}, emptyFinalizer)
	clusterRole5                    = newClusterRole("gitjob-clusterRole5", map[string]string{"clusterRole5": "true"}, emptyFinalizer)
	clusterRoleBinding6             = newClusterRoleBinding("pod-impersonation-helm-clusterRoleBinding6", map[string]string{"clusterRoleBinding6": "true"}, emptyFinalizer)
	clusterRole6                    = newClusterRole("pod-impersonation-helm-clusterRole6", map[string]string{"clusterRole6": "true"}, emptyFinalizer)
	podSecurityPolicy1              = newPodSecurityPolicy("podSecurityPolicy1", map[string]string{"app.kubernetes.io/name": "rancher-logging", "podSecurityPolicy1": "true"})
	podSecurityPolicy2              = newPodSecurityPolicy("rancher-logging-rke-aggregator", map[string]string{"podSecurityPolicy2": "true"})
	podSecurityPolicy3              = newPodSecurityPolicy("podSecurityPolicy3", map[string]string{"release": "rancher-monitoring", "podSecurityPolicy3": "true"})
	podSecurityPolicy4              = newPodSecurityPolicy("podSecurityPolicy4", map[string]string{"app": "rancher-monitoring-crd-manager", "podSecurityPolicy4": "true"})
	podSecurityPolicy5              = newPodSecurityPolicy("podSecurityPolicy5", map[string]string{"app": "rancher-monitoring-patch-sa", "podSecurityPolicy5": "true"})
	podSecurityPolicy6              = newPodSecurityPolicy("podSecurityPolicy6", map[string]string{"app.kubernetes.io/instance": "rancher-monitoring", "podSecurityPolicy6": "true"})
	podSecurityPolicy7              = newPodSecurityPolicy("podSecurityPolicy7", map[string]string{"release": "rancher-gatekeeper", "podSecurityPolicy7": "true"})
	podSecurityPolicy8              = newPodSecurityPolicy("podSecurityPolicy8", map[string]string{"app": "rancher-gatekeeper-crd-manager", "podSecurityPolicy8": "true"})
	podSecurityPolicy9              = newPodSecurityPolicy("podSecurityPolicy9", map[string]string{"app.kubernetes.io/name": "rancher-backup", "podSecurityPolicy9": "true"})
)

// Test_rancherUninstall - test the uninstall of Rancher
func Test_rancherUninstall(t *testing.T) {
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
	verify(t, ctx, fakeDynamicClient, false)

	// Call the function being tested
	cleanupPreventRecreate(ctx)
	cleanupWebhooks(ctx)
	cleanupClusterRolesAndBindings(ctx)
	cleanupPodSecurityPolicies(ctx)
	cleanupNamespaces(ctx)

	// Verify expected resources do not exist after  the cleanup
	verify(t, ctx, fakeDynamicClient, true)
}

// verify - verify expected counts of resources before and after the rancher cleanup
func verify(t *testing.T, ctx spi.ComponentContext, fakeDynamicClient dynamic.Interface, cleanupDone bool) {
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

	verifyResources(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: "v1", Resource: "mutatingwebhookconfigurations"}, []string{cattleNameFilter, webhookMonitorFilter}, expectedLen)
	verifyResources(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: "v1", Resource: "validatingwebhookconfigurations"}, []string{cattleNameFilter, webhookMonitorFilter}, expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, normanSelector, expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, normanSelector, expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, "clusterRoleBinding2=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, "clusterRole2=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, "clusterRoleBinding3=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, "clusterRole3=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, "clusterRoleBinding4=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, "clusterRole4=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, "clusterRoleBinding5=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, "clusterRole5=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, "clusterRoleBinding6=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, "clusterRole6=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"}, "podSecurityPolicy1=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"}, "podSecurityPolicy2=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"}, "podSecurityPolicy3=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"}, "podSecurityPolicy4=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"}, "podSecurityPolicy5=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"}, "podSecurityPolicy6=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"}, "podSecurityPolicy7=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"}, "podSecurityPolicy8=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"}, "podSecurityPolicy9=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace1=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace2=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace3=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace4=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace5=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace6=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace7=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace8=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace9=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace10=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace11=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace12=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace13=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace14=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace15=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace16=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace17=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace18=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace19=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace20=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace21=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace22=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace23=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace24=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace25=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace26=true", expectedLen)
	verifyResource(t, ctx, fakeDynamicClient, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, "namespace27=true", expectedLen)
}

func verifyResources(t *testing.T, ctx spi.ComponentContext, fakeDynamicClient dynamic.Interface, gvr schema.GroupVersionResource, nameFilter []string, expectedLen int) {
	list, err := listResource(ctx, fakeDynamicClient, gvr, "")
	assert.NoError(t, err)
	for _, item := range list.Items {
		if expectedLen == 0 {
			for _, name := range nameFilter {
				assert.NotContains(t, item.GetName(), name)
			}
		} else {
			found := false
			for _, name := range nameFilter {
				if strings.Contains(item.GetName(), name) {
					found = true
					break
				}
			}
			assert.True(t, found)
		}
	}
}
func verifyResource(t *testing.T, ctx spi.ComponentContext, fakeDynamicClient dynamic.Interface, gvr schema.GroupVersionResource, labelSelector string, expectedLen int) {
	list, err := listResource(ctx, fakeDynamicClient, gvr, labelSelector)
	assert.NoError(t, err)
	assert.Equal(t, expectedLen, len(list.Items))
}

func getSchemeForCleanup() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = admv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	_ = policyv1.AddToScheme(scheme)
	return scheme
}

func newClusterCleanupRepoResources() []runtime.Object {
	return []runtime.Object{deployment, daemonSet, mutatingWebhookConfiguration, validatingWebhookConfiguration,
		mutatingWebhookConfiguration2, validatingWebhookConfiguration2, clusterRoleBinding1, clusterRole1,
		clusterRole2, clusterRoleBinding2, clusterRole3, clusterRoleBinding3, clusterRole4, clusterRoleBinding4,
		clusterRole5, clusterRoleBinding5, clusterRole6, clusterRoleBinding6,
		podSecurityPolicy1, podSecurityPolicy2, podSecurityPolicy3, podSecurityPolicy4, podSecurityPolicy5,
		podSecurityPolicy6, podSecurityPolicy7, podSecurityPolicy8, podSecurityPolicy9,
		namespace1, namespace2, namespace3, namespace4, namespace5, namespace6, namespace7, namespace8, namespace9,
		namespace10, namespace11, namespace12, namespace13, namespace14, namespace15, namespace16, namespace17, namespace18, namespace19,
		namespace20, namespace21, namespace22, namespace23, namespace24, namespace25, namespace26, namespace27}
}

func newDeployment(namespace string, name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func newDaemonSet(namespace string, name string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func newPodSecurityPolicy(name string, labels map[string]string) *policyv1.PodSecurityPolicy {
	return &policyv1.PodSecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func newValidatingWebhookConfiguration(name string) *admv1.ValidatingWebhookConfiguration {
	return &admv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func newMutatingWebhookConfiguration(name string) *admv1.MutatingWebhookConfiguration {
	return &admv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}
