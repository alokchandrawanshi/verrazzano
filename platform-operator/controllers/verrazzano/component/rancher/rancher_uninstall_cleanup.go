// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

import (
	"context"

	"k8s.io/client-go/dynamic"

	"github.com/verrazzano/verrazzano/pkg/k8sutil"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// getDynamicClientForCleanupFunc is the function for getting a k8s dynamic client - this allows us to override
// the function for unit testing
var getDynamicClientForCleanupFunc getDynamicClientFuncSig = getDynamicClientForCleanup

// cleanupPreventRecreate - Implement the following portion of rancher-cleanup in golang
//
//	# Delete rancher install to not have anything running that (re)creates resources
//	kcd "-n cattle-system deploy,ds --all"
//	kubectl -n cattle-system wait --for delete pod --selector=app=rancher
//	# Delete the only resource not in cattle namespaces
//	kcd "-n kube-system configmap cattle-controllers"
func cleanupPreventRecreate(ctx spi.ComponentContext) {
	// Delete rancher install to not have anything running that (re)creates resources
	deleteAll(ctx, "apps", "v1", "Deployment", ComponentNamespace)
	deleteAll(ctx, "apps", "v1", "DaemonSet", ComponentNamespace)
}

func deleteAll(ctx spi.ComponentContext, group string, version string, resource string, namespace string) {
	dynClient, err := getDynamicClientForCleanupFunc()
	if err != nil {
		ctx.Log().Errorf("Component %s failed to get dynamic client: %v", ComponentName, err)
		return
	}
	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	list, err := dynClient.Resource(resourceId).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		ctx.Log().Errorf("Component %s failed to list Deployments in namespace %s: %v", ComponentName, ComponentNamespace, err)
		return
	}

	// Delete each of the items returned
	for _, item := range list.Items {
		resourceId = schema.GroupVersionResource{
			Group:    item.GroupVersionKind().Group,
			Version:  item.GroupVersionKind().Version,
			Resource: item.GroupVersionKind().Kind,
		}
		err = dynClient.Resource(resourceId).Namespace(item.GetNamespace()).Delete(context.TODO(), item.GetName(), metav1.DeleteOptions{})
		if err != nil {
			ctx.Log().Errorf("Component %s failed to delete %s %s/%s: %v", resourceId.Resource, item.GetNamespace(), item.GetName(), err)
		}
	}
}

func getDynamicClientForCleanup() (dynamic.Interface, error) {
	dynClient, err := k8sutil.GetDynamicClient()
	if err != nil {
		return nil, err
	}
	return dynClient, nil
}
