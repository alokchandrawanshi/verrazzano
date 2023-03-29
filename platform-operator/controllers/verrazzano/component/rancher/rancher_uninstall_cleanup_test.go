// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/stretchr/testify/assert"
	vzapi "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/common"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	appsv1 "k8s.io/api/apps/v1"
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
	daemonset = &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: common.CattleSystem,
			Name:      common.RancherName,
		},
	}
)

func Test_deleteAll(t *testing.T) {
	type args struct {
		group     string
		version   string
		resource  string
		namespace string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "delete deployments",
			args: args{
				group:     "apps",
				version:   "v1",
				resource:  "deployments",
				namespace: ComponentNamespace,
			},
		},
		{
			name: "delete daemonsets",
			args: args{
				group:     "apps",
				version:   "v1",
				resource:  "daemonsets",
				namespace: ComponentNamespace,
			},
		},
	}
	client := fake.NewClientBuilder().WithScheme(getSchemeForCleanup()).Build()
	ctx := spi.NewFakeContext(client, &vzapi.Verrazzano{}, nil, false)

	// create a fake dynamic client to serve the Setting and ClusterRepo resources
	fakeDynamicClient := dynfake.NewSimpleDynamicClient(getSchemeForCleanup(), newClusterCleanupRepoResources()...)

	// override the getDynamicClientFunc for unit testing and reset it when done
	prevGetDynamicClientFunc := getDynamicClientForCleanupFunc
	getDynamicClientForCleanupFunc = func() (dynamic.Interface, error) { return fakeDynamicClient, nil }
	defer func() {
		getDynamicClientForCleanupFunc = prevGetDynamicClientFunc
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteAll(ctx, tt.args.group, tt.args.version, tt.args.resource, tt.args.namespace)
			gvr := schema.GroupVersionResource{Group: tt.args.group, Version: tt.args.version, Resource: tt.args.resource}
			list, err := fakeDynamicClient.Resource(gvr).Namespace(tt.args.namespace).List(context.TODO(), metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Equal(t, 0, len(list.Items))
		})
	}
}

func getSchemeForCleanup() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	return scheme
}

func newClusterCleanupRepoResources() []runtime.Object {
	return []runtime.Object{deployment, daemonset}
}
