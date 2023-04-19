// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

import (
	"context"
	"strings"

	"k8s.io/client-go/discovery"

	"github.com/verrazzano/verrazzano/pkg/k8sutil"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var cattleNameFilter = "cattle.io"
var webhookMonitorFilter = "rancher-monitoring"
var normanSelector = "cattle.io/creator=norman"

// getDynamicClientForCleanupFunc is the function for getting a k8s dynamic client - this allows us to override
// the function for unit testing
var getDynamicClientForCleanupFunc getDynamicClientFuncSig = getDynamicClientForCleanup

// deleteOptions - filter settings for a delete resources request
type deleteOptions struct {
	Namespace              string
	RemoveCattleFinalizers bool
	LabelSelector          string
	NameFilter             []string
	NameMatchType          nameMatchType
}

type nameMatchType string

const (
	Contains  nameMatchType = "contains"
	Equals    nameMatchType = "equals"
	HasPrefix nameMatchType = "startsWith"
)

// defaultDeleteOptions - create an instance of deleteOptions with default values
func defaultDeleteOptions() deleteOptions {
	return deleteOptions{
		RemoveCattleFinalizers: false,
		LabelSelector:          "",
		NameFilter:             []string{},
		NameMatchType:          Contains,
	}
}

// cleanupRancher - perform the functions of the rancher-cleanup job
func cleanupRancher(ctx spi.ComponentContext) {
	cleanupPreventRecreate(ctx)
	cleanupWebhooks(ctx)
	cleanupClusterRolesAndBindings(ctx)
	cleanupPodSecurityPolicies(ctx)
	cleanupApiResources(ctx)
	cleanupNamespaces(ctx)
}

// cleanupPreventRecreate - delete resources that would recreate resources during the cleanup
func cleanupPreventRecreate(ctx spi.ComponentContext) {
	options := defaultDeleteOptions()
	options.Namespace = ComponentNamespace
	deleteResources(ctx, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, options)
	deleteResources(ctx, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}, options)
}

// cleanupWebhooks - Implement the portion of rancher-cleanup script that deletes webhooks
func cleanupWebhooks(ctx spi.ComponentContext) {
	deleteResources(ctx, schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: "v1", Resource: "mutatingwebhookconfigurations"}, defaultDeleteOptions())
	deleteResources(ctx, schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: "v1", Resource: "validatingwebhookconfigurations"}, defaultDeleteOptions())
}

// cleanupClusterRolesAndBindings - Implement the portion of the rancher-cleanup script that deletes ClusterRoles and ClusterRoleBindings
func cleanupClusterRolesAndBindings(ctx spi.ComponentContext) {
	options := defaultDeleteOptions()
	options.RemoveCattleFinalizers = true

	options.LabelSelector = normanSelector
	deleteResources(ctx, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, options)
	deleteResources(ctx, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, options)
	options.LabelSelector = ""

	options.NameFilter = []string{"rancher"}
	options.NameMatchType = Contains
	deleteResources(ctx, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, options)
	deleteResources(ctx, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, options)

	options.NameFilter = []string{"cattle-", "fleet-", "gitjob", "pod-impersonation-helm-"}
	options.NameMatchType = HasPrefix
	deleteResources(ctx, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"}, options)
	deleteResources(ctx, schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"}, options)
}

// cleanupPodSecurityPolicies - Implement the portion of the rancher-cleanup script that deletes PodSecurityPolicies
func cleanupPodSecurityPolicies(ctx spi.ComponentContext) {
	// Delete policies by label selector
	labelSelectors := []string{"app.kubernetes.io/name=rancher-logging", "release=rancher-monitoring", "app=rancher-monitoring-crd-manager",
		"app=rancher-monitoring-patch-sa", "app.kubernetes.io/instance=rancher-monitoring", "release=rancher-gatekeeper",
		"app=rancher-gatekeeper-crd-manager", "app.kubernetes.io/name=rancher-backup"}
	options := defaultDeleteOptions()
	for _, selector := range labelSelectors {
		options.LabelSelector = selector
		deleteResources(ctx, schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"}, options)
	}

	options.LabelSelector = ""
	options.NameFilter = []string{"rancher-logging-rke-aggregator"}
	options.NameMatchType = Equals
	deleteResources(ctx, schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"}, options)
}

// cleanupApiResources - Implement the portion of the rancher-cleanup script that deletes API Resources
func cleanupApiResources(ctx spi.ComponentContext) {

	config, err := k8sutil.GetConfigFromController()
	if err != nil {
		return
	}
	disco, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return
	}
	lists, err := disco.ServerPreferredResources()
	if err != nil {
		return
	}
	for _, list := range lists {
		for _, resource := range list.APIResources {
			// Skip namespaced resources
			if resource.Namespaced {
				continue
			}

			// Skip resources that do not support delete verb
			if len(resource.Verbs) == 0 {
				continue
			}
			if !strings.Contains(resource.Verbs.String(), "delete") {
				continue
			}

			// Create a list of all resources that match the delete filter
			if strings.Contains(resource.Name, cattleNameFilter) {
				ctx.Log().Infof("Resource type %s/%s/%s with name %s should be deleted", resource.Group, resource.Version, resource.Kind, resource.Name)
			}
		}
	}
}

// cleanupNamespaces - Implement the portion of the rancher-cleanup script that deletes namespaces
func cleanupNamespaces(ctx spi.ComponentContext) {
	options := defaultDeleteOptions()

	// Cattle namespaces
	options.NameFilter = []string{"local", "cattle-system", "cattle-impersonation-system", "cattle-global-data", "cattle-global-nt"}
	deleteResources(ctx, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, options)

	// Tools namespaces
	options.NameFilter = []string{"cattle-resources-system", "cis-operator-system", "cattle-dashboards", "cattle-gatekeeper-system", "cattle-alerting",
		"cattle-logging", "cattle-pipeline", "cattle-prometheus", "rancher-operator-system", "cattle-monitoring-system", "cattle-logging-system"}
	deleteResources(ctx, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, options)

	// Fleet namespaces
	options.NameFilter = []string{"cattle-fleet-clusters-system", "cattle-fleet-local-system", "cattle-fleet-system", "fleet-default", "fleet-local", "fleet-system"}
	deleteResources(ctx, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, options)

	// Prefixed namespaces
	options.NameFilter = []string{"cluster-fleet", "p-", "c-", "user-", "u-"}
	options.NameMatchType = HasPrefix
	deleteResources(ctx, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}, options)
}

// deleteResources - Delete all instances of a resource that meet the filters passed
func deleteResources(ctx spi.ComponentContext, gvr schema.GroupVersionResource, options deleteOptions) {
	var errorList []error
	dynClient, err := getClient(ctx)
	if err != nil {
		return
	}

	var list *unstructured.UnstructuredList
	if len(options.Namespace) > 0 {
		list, err = listResourceByNamespace(ctx, dynClient, gvr, options.Namespace, options.LabelSelector)
	} else {
		list, err = listResource(ctx, dynClient, gvr, options.LabelSelector)
	}
	if err != nil {
		return
	}

	// Delete each of the items returned
	for i, item := range list.Items {
		if options.RemoveCattleFinalizers {
			err = removeFinalizer(ctx, &list.Items[i], []string{finalizerSubString})
			if err != nil {
				errorList = append(errorList, err)
			}
		}
		if len(options.NameFilter) == 0 {
			deleteResource(ctx, dynClient, gvr, item)
		} else {
			for _, filter := range options.NameFilter {
				if options.NameMatchType == Contains && strings.Contains(item.GetName(), filter) {
					deleteResource(ctx, dynClient, gvr, item)
				} else if options.NameMatchType == HasPrefix && strings.HasPrefix(item.GetName(), filter) {
					deleteResource(ctx, dynClient, gvr, item)
				} else if options.NameMatchType == Equals && strings.EqualFold(item.GetName(), filter) {
					deleteResource(ctx, dynClient, gvr, item)
				}
			}
		}
	}
}

// deleteResource - delete a single instance of a resource
func deleteResource(ctx spi.ComponentContext, dynClient dynamic.Interface, gvr schema.GroupVersionResource, item unstructured.Unstructured) {
	err := dynClient.Resource(gvr).Namespace(item.GetNamespace()).Delete(context.TODO(), item.GetName(), metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		ctx.Log().Errorf("Component %s failed to delete %s %s/%s: %v", gvr.Resource, item.GetNamespace(), item.GetName(), err)
	}
}

// listResource - common function to list resource without a Namespace
func listResource(ctx spi.ComponentContext, dynClient dynamic.Interface, gvr schema.GroupVersionResource, labelSelector string) (*unstructured.UnstructuredList, error) {
	listOptions := metav1.ListOptions{}
	listOptions.LabelSelector = labelSelector
	list, err := dynClient.Resource(gvr).List(context.TODO(), listOptions)
	if err != nil {
		ctx.Log().Errorf("Component %s failed to list %s: %v", ComponentName, gvr.Resource, err)
		return nil, err
	}
	return list, nil
}

// listResourceByNamespace - common function for listing resources
func listResourceByNamespace(ctx spi.ComponentContext, dynClient dynamic.Interface, gvr schema.GroupVersionResource, namespace string, labelSelector string) (*unstructured.UnstructuredList, error) {
	listOptions := metav1.ListOptions{}
	listOptions.LabelSelector = labelSelector
	list, err := dynClient.Resource(gvr).Namespace(namespace).List(context.TODO(), listOptions)
	if err != nil {
		ctx.Log().Errorf("Component %s failed to list %s/%s: %v", ComponentName, ComponentNamespace, gvr.Resource, err)
		return nil, err
	}
	return list, nil
}

// getClient - common function to get a dynamic client and log any error that occurs
func getClient(ctx spi.ComponentContext) (dynamic.Interface, error) {
	dynClient, err := getDynamicClientForCleanupFunc()
	if err != nil {
		ctx.Log().Errorf("Component %s failed to get dynamic client: %v", ComponentName, err)
		return nil, err
	}
	return dynClient, nil
}

// getDynamicClientForCleanup - return a dynamic client, this function may be overridden for unit testing
func getDynamicClientForCleanup() (dynamic.Interface, error) {
	dynClient, err := k8sutil.GetDynamicClient()
	if err != nil {
		return nil, err
	}
	return dynClient, nil
}
