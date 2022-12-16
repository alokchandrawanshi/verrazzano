// Copyright (c) 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package webhooks

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	"github.com/verrazzano/verrazzano/tools/vz/pkg/helpers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	ctrlfakes "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
)

// TestHandleAnnotateMysqlBackupJob tests handling an admission.Request
// GIVEN a MysqlBackupJobWebhook and an admission.Request
// WHEN Handle is called with an admission.Request containing job created by mysql-operator
// THEN Handle should return an Allowed response with no action required
func TestPodSecurityHandle(t *testing.T) {

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podSecurityConfigMapName,
			Namespace: constants.VerrazzanoInstallNamespace,
		},
		Data: map[string]string{
			ignoredNamespacesKey: "",
			ignoredPodsKey:       "",
		},
	}

	client := ctrlfakes.NewClientBuilder().WithScheme(helpers.NewScheme()).WithObjects(cm).Build()

	defaulter := &PodSecurityWebhook{
		DynamicClient: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
		KubeClient:    fake.NewSimpleClientset(),
		Client:        client,
	}

	// Create a pod with Istio injection disabled
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-ns",
		},
		Spec: corev1.PodSpec{
			Containers:     []corev1.Container{{}},
			InitContainers: []corev1.Container{{}},
		},
	}

	pod, err := defaulter.KubeClient.CoreV1().Pods("test-ns").Create(context.TODO(), pod, metav1.CreateOptions{})
	assert.NoError(t, err, "Unexpected error creating pod")
	decoder := decoder()
	err = defaulter.InjectDecoder(decoder)
	assert.NoError(t, err, "Unexpected error injecting decoder")
	req := admission.Request{}
	req.Namespace = "test-ns"
	marshaledPod, err := json.Marshal(pod)
	assert.NoError(t, err, "Unexpected error marshaling pod")
	req.Object = runtime.RawExtension{Raw: marshaledPod}
	res := defaulter.Handle(context.TODO(), req)
	assert.True(t, res.Allowed)
	assert.NoError(t, err, "Unexpected error marshaling pod")
	assert.Equal(t, "/spec/initContainers/0/securityContext", res.Patches[0].Path)
	assert.Equal(t, "/spec/containers/0/securityContext", res.Patches[1].Path)
}

func TestPodSecurityHandleIgnorePod(t *testing.T) {

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podSecurityConfigMapName,
			Namespace: constants.VerrazzanoInstallNamespace,
		},
		Data: map[string]string{
			ignoredNamespacesKey: "",
			ignoredPodsKey: `
- namespace: test-ns
  name: test-pod`,
		},
	}

	client := ctrlfakes.NewClientBuilder().WithScheme(helpers.NewScheme()).WithObjects(cm).Build()

	defaulter := &PodSecurityWebhook{
		DynamicClient: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
		KubeClient:    fake.NewSimpleClientset(),
		Client:        client,
	}

	// Create a pod with Istio injection disabled
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-ns",
		},
		Spec: corev1.PodSpec{
			Containers:     []corev1.Container{{}},
			InitContainers: []corev1.Container{{}},
		},
	}

	pod, err := defaulter.KubeClient.CoreV1().Pods("test-ns").Create(context.TODO(), pod, metav1.CreateOptions{})
	assert.NoError(t, err, "Unexpected error creating pod")
	decoder := decoder()
	err = defaulter.InjectDecoder(decoder)
	assert.NoError(t, err, "Unexpected error injecting decoder")
	req := admission.Request{}
	req.Namespace = "test-ns"
	marshaledPod, err := json.Marshal(pod)
	assert.NoError(t, err, "Unexpected error marshaling pod")
	req.Object = runtime.RawExtension{Raw: marshaledPod}
	res := defaulter.Handle(context.TODO(), req)
	assert.True(t, res.Allowed)
	assert.NoError(t, err, "Unexpected error marshaling pod")
	assert.Equal(t, 0, len(res.Patches))
}
