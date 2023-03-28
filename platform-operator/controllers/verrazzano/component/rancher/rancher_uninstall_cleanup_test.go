// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package rancher

import (
	"testing"

	vzapi "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	deployment = &appsv1.Deployment{
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
			name: "foo",
			args: args{
				group:     "apps",
				version:   "v1",
				resource:  "Deployment",
				namespace: ComponentNamespace,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewClientBuilder().WithScheme(getSchemeForCleanup()).WithObjects(deployment).Build()
			ctx := spi.NewFakeContext(client, &vzapi.Verrazzano{}, nil, false)

			deleteAll(ctx, tt.args.group, tt.args.version, tt.args.resource, tt.args.namespace)
		})
	}
}

func getSchemeForCleanup() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	return scheme
}
