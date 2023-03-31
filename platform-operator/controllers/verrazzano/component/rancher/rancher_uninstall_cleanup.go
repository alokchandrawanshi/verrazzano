// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

import (
	"context"
	"strings"

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

// getDynamicClientForCleanupFunc is the function for getting a k8s dynamic client - this allows us to override
// the function for unit testing
var getDynamicClientForCleanupFunc getDynamicClientFuncSig = getDynamicClientForCleanup

// cleanupRancher - perform the functions of the rancher-cleanup job
func cleanupRancher(ctx spi.ComponentContext) {
	cleanupPreventRecreate(ctx)
	cleanupWebhooks(ctx)
}

// cleanupPreventRecreate - delete resources that would recreate resources during the cleanup
func cleanupPreventRecreate(ctx spi.ComponentContext) {
	// Delete rancher install to not have anything running that (re)creates resources
	deleteResources(ctx, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, ComponentNamespace, []string{})
	deleteResources(ctx, schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}, ComponentNamespace, []string{})
}

// cleanupWebhooks - Implement the portion of rancher-cleanup script that deletes webhooks
func cleanupWebhooks(ctx spi.ComponentContext) {
	nameFilter := []string{cattleNameFilter, webhookMonitorFilter}
	// Delete any blocking webhooks from preventing requests
	deleteResources(ctx, schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: "v1", Resource: "mutatingwebhookconfigurations"}, "", nameFilter)
	deleteResources(ctx, schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: "v1", Resource: "validatingwebhookconfigurations"}, "", nameFilter)
}

// cleanupClusterRolesAndBindings - Implement the portion of the rancher-cleanup script
// that deletes ClusterRoles and ClusterRoleBindings
func cleanupClusterRolesAndBindings(ctx spi.ComponentContext) {

}

// deleteResources - Delete all instances of a resource that meet the filters passed
func deleteResources(ctx spi.ComponentContext, resourceId schema.GroupVersionResource, namespace string, nameFilter []string) {
	dynClient, err := getClient(ctx)
	if err != nil {
		return
	}

	var list *unstructured.UnstructuredList
	if len(namespace) > 0 {
		list, err = listResourceByNamespace(ctx, dynClient, resourceId, namespace)
	} else {
		list, err = listResource(ctx, dynClient, resourceId)
	}
	if err != nil {
		return
	}

	// Delete each of the items returned
	for _, item := range list.Items {
		if len(nameFilter) == 0 {
			deleteResource(ctx, dynClient, resourceId, item)
		} else {
			for _, filter := range nameFilter {
				if strings.Contains(item.GetName(), filter) {
					deleteResource(ctx, dynClient, resourceId, item)
				}
			}
		}
	}
}

// deleteResource - delete a single instance of a resource
func deleteResource(ctx spi.ComponentContext, dynClient dynamic.Interface, resourceId schema.GroupVersionResource, item unstructured.Unstructured) {
	err := dynClient.Resource(resourceId).Namespace(item.GetNamespace()).Delete(context.TODO(), item.GetName(), metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		ctx.Log().Errorf("Component %s failed to delete %s %s/%s: %v", resourceId.Resource, item.GetNamespace(), item.GetName(), err)
	}
}

// listResource - common function to list resource without a namespace
func listResource(ctx spi.ComponentContext, dynClient dynamic.Interface, resourceId schema.GroupVersionResource) (*unstructured.UnstructuredList, error) {
	list, err := dynClient.Resource(resourceId).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		ctx.Log().Errorf("Component %s failed to list %s: %v", ComponentName, resourceId.Resource, err)
		return nil, err
	}
	return list, nil
}

// listResourceByNamespace - common function for listing resources
func listResourceByNamespace(ctx spi.ComponentContext, dynClient dynamic.Interface, resourceId schema.GroupVersionResource, namespace string) (*unstructured.UnstructuredList, error) {
	list, err := dynClient.Resource(resourceId).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		ctx.Log().Errorf("Component %s failed to list %s/%s: %v", ComponentName, ComponentNamespace, resourceId.Resource, err)
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
