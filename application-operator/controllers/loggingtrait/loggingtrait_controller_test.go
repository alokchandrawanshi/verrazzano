// Copyright (c) 2021, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package loggingtrait

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	vzapi "github.com/verrazzano/verrazzano/application-operator/apis/oam/v1alpha1"
	"github.com/verrazzano/verrazzano/application-operator/mocks"
	vzconst "github.com/verrazzano/verrazzano/pkg/constants"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"

	oamrt "github.com/crossplane/crossplane-runtime/apis/common/v1"
	oamcore "github.com/crossplane/oam-kubernetes-runtime/apis/core/v1alpha2"
	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	asserts "github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	k8sapps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8smeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// +kubebuilder:scaffold:imports
)

var (
	namespaceName               = "test-namespace"
	workloadName                = "test-workload-name"
	workloadUID                 = "test-workload-uid"
	traitName                   = "test-trait-name"
	deploymentName              = "test-deployment-name"
	workloadDefinitionNamespace = "containerizedworkloads.core.oam.dev"
	serverErr                   = "server error"
)

func TestReconcilerSetupWithManager(t *testing.T) {
	assert := asserts.New(t)

	var mocker *gomock.Controller
	var mgr *mocks.MockManager
	var cli *mocks.MockClient
	var scheme *runtime.Scheme

	var reconciler LoggingTraitReconciler
	var err error

	mocker = gomock.NewController(t)
	mgr = mocks.NewMockManager(mocker)
	cli = mocks.NewMockClient(mocker)
	scheme = runtime.NewScheme()
	_ = vzapi.AddToScheme(scheme)
	reconciler = LoggingTraitReconciler{Client: cli, Scheme: scheme}
	mgr.EXPECT().GetControllerOptions().AnyTimes()
	mgr.EXPECT().GetScheme().Return(scheme)
	mgr.EXPECT().GetLogger().Return(logr.Discard())
	mgr.EXPECT().SetFields(gomock.Any()).Return(nil).AnyTimes()
	mgr.EXPECT().Add(gomock.Any()).Return(nil).AnyTimes()
	err = reconciler.SetupWithManager(mgr)
	mocker.Finish()
	assert.NoError(err)
}

// TestLoggingTraitCreatedForContainerizedWorkload tests the creation of a logging trait related to a containerized workload.
// GIVEN a logging trait that has been created
// AND the logging trait is related to a containerized workload
// WHEN the logging trait Reconcile method is invoked
// THEN verify that logging trait finalizer is added
// AND verify that pod annotations are updated
// AND verify that the scraper configmap is updated
// AND verify that the scraper pod is restarted
func TestLoggingTraitCreatedForContainerizedWorkload(t *testing.T) {
	assert := asserts.New(t)
	mocker := gomock.NewController(t)
	mock := mocks.NewMockClient(mocker)
	mockStatus := mocks.NewMockStatusWriter(mocker)

	testDeployment := newDeployment(deploymentName, namespaceName, workloadName, workloadUID)

	// Expect a call to get the logging trait
	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(types.NamespacedName{Namespace: namespaceName, Name: traitName}), gomock.Not(gomock.Nil()), gomock.Any()).
		DoAndReturn(func(ctx context.Context, name types.NamespacedName, trait *vzapi.LoggingTrait, opt ...client.GetOption) error {
			trait.SetWorkloadReference(oamrt.TypedReference{
				APIVersion: oamcore.SchemeGroupVersion.Identifier(),
				Kind:       oamcore.ContainerizedWorkloadKind,
				Name:       workloadName,
				UID:        types.UID(workloadUID),
			})
			trait.SetNamespace(namespaceName)
			return nil
		})
	// Expect a call to get the workload
	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(client.ObjectKey{Namespace: namespaceName, Name: workloadName}), gomock.Not(gomock.Nil()), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key client.ObjectKey, workload *unstructured.Unstructured, opt ...client.GetOption) error {
			return nil
		})
	// Expect a call to get the workload definition
	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(types.NamespacedName{Namespace: "", Name: workloadDefinitionNamespace}), gomock.Not(gomock.Nil()), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key client.ObjectKey, workloadDef *oamcore.WorkloadDefinition, opt ...client.GetOption) error {
			workloadDef.Spec.ChildResourceKinds = []oamcore.ChildResourceKind{
				{
					APIVersion: k8sapps.SchemeGroupVersion.Identifier(),
					Kind:       "Deployment",
				},
			}
			return nil
		})
	// Expect to list config map
	options := []client.ListOption{client.InNamespace(namespaceName), client.MatchingFields{"metadata.name": "logging-stdout-test-deployment-name-deployment"}}
	mock.EXPECT().
		List(gomock.Any(), gomock.Not(gomock.Nil()), options).
		DoAndReturn(func(ctx context.Context, list *unstructured.UnstructuredList, opts ...client.ListOption) error {
			return nil
		})
	// Expect to create a config map
	mock.EXPECT().
		Create(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any()).
		DoAndReturn(func(ctx context.Context, configMap *corev1.ConfigMap, opts ...client.CreateOption) error {
			return nil
		})
	// Expect a call to get the deployment
	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(client.ObjectKey{Namespace: namespaceName, Name: deploymentName}), gomock.Not(gomock.Nil()), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key client.ObjectKey, workload *unstructured.Unstructured, opt ...client.GetOption) error {
			return nil
		})
	// Expect a call to list the child Deployment resources of the containerized workload definition
	mock.EXPECT().
		List(gomock.Any(), gomock.Not(gomock.Nil()), gomock.Any()).
		DoAndReturn(func(ctx context.Context, list *unstructured.UnstructuredList, opts ...client.ListOption) error {
			assert.Equal("Deployment", list.GetKind())
			return appendAsUnstructured(list, testDeployment)
		})
	// Expect a call to get the status writer
	mock.EXPECT().Status().Return(mockStatus).AnyTimes()

	// Create and make the request
	request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: namespaceName, Name: "test-trait-name"}}

	reconciler := newLoggingTraitReconciler(mock, t)
	result, err := reconciler.Reconcile(context.TODO(), request)

	// Validate the results
	mocker.Finish()
	assert.NoError(err)
	assert.Equal(time.Duration(0), result.RequeueAfter)
}

// TestDeleteLoggingTraitFromContainerizedWorkload tests the deletion of a logging trait related to a containerized workload.
// GIVEN a logging trait
// AND the logging trait is related to a containerized workload
// WHEN the logging trait reconcileTraitDelete method is invoked
// THEN verify that the logging trait has been deleted
func TestDeleteLoggingTraitFromContainerizedWorkload(t *testing.T) {
	assert := asserts.New(t)
	mocker := gomock.NewController(t)
	mock := mocks.NewMockClient(mocker)

	testDeployment := newDeployment(deploymentName, namespaceName, workloadName, workloadUID)

	// Create trait for deletion
	trait := vzapi.LoggingTrait{
		TypeMeta: k8smeta.TypeMeta{
			Kind: vzapi.LoggingTraitKind,
		},
		ObjectMeta: k8smeta.ObjectMeta{
			Name:      traitName,
			Namespace: namespaceName,
		},
		Spec: vzapi.LoggingTraitSpec{
			WorkloadReference: oamrt.TypedReference{
				APIVersion: oamcore.SchemeGroupVersion.Identifier(),
				Kind:       oamcore.ContainerizedWorkloadKind,
				Name:       workloadName,
				UID:        types.UID(workloadUID),
			},
		},
		Status: vzapi.LoggingTraitStatus{},
	}

	// Expect a call to get the workload
	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(types.NamespacedName{Namespace: namespaceName, Name: workloadName}), gomock.Not(gomock.Nil()), gomock.Any()).
		DoAndReturn(func(ctx context.Context, name types.NamespacedName, workload *unstructured.Unstructured, opt ...client.GetOption) error {
			workload.SetNamespace(namespaceName)
			return nil
		})
	// Expect a call to get the workload definition
	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(types.NamespacedName{Namespace: "", Name: workloadDefinitionNamespace}), gomock.Not(gomock.Nil()), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key client.ObjectKey, workloadDef *oamcore.WorkloadDefinition, opt ...client.GetOption) error {
			workloadDef.Spec.ChildResourceKinds = []oamcore.ChildResourceKind{
				{
					APIVersion: k8sapps.SchemeGroupVersion.Identifier(),
					Kind:       "Deployment",
				},
			}
			return nil
		})
	// Expect to list deployment
	options := []client.ListOption{client.InNamespace(namespaceName)}
	mock.EXPECT().
		List(gomock.Any(), gomock.Not(gomock.Nil()), options).
		DoAndReturn(func(ctx context.Context, deployment *unstructured.UnstructuredList, opts ...client.ListOption) error {
			unstructuredDeployment, err := convertToUnstructured(testDeployment)
			if err != nil {
				t.Fatalf("Could not create unstructured Deployment")
			}
			deployment.Items = []unstructured.Unstructured{unstructuredDeployment}
			return nil
		})
	// Expect a call to get the deployment
	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(client.ObjectKey{Namespace: namespaceName, Name: deploymentName}), gomock.Not(gomock.Nil()), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key client.ObjectKey, workload *unstructured.Unstructured, opt ...client.GetOption) error {
			return nil
		})
	// Expect to list config map
	mock.EXPECT().
		List(gomock.Any(), gomock.Not(gomock.Nil()), client.InNamespace(namespaceName), client.MatchingFields{"metadata.name": "logging-stdout-test-deployment-name-deployment"}).
		DoAndReturn(func(ctx context.Context, list *unstructured.UnstructuredList, options ...client.ListOption) error {
			return nil
		})

	reconciler := newLoggingTraitReconciler(mock, t)
	result, err := reconciler.reconcileTraitDelete(context.TODO(), vzlog.DefaultLogger(), &trait)

	// Validate the results
	mocker.Finish()
	assert.NoError(err)
	assert.Equal(time.Duration(0), result.RequeueAfter)
}

// Test_fetchTrait tests the fetchTrait function of the LoggingTraitReconciler
// GIVEN a call to fetchTrait method of the LoggingTrait Reconciler
// WHEN there is some error during retrieving the trait
// THEN expect the reconciler to requeue and return no error
func TestFetchTraitError(t *testing.T) {
	assert := asserts.New(t)
	mocker := gomock.NewController(t)
	mock := mocks.NewMockClient(mocker)

	request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: namespaceName, Name: traitName}}

	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(types.NamespacedName{Namespace: namespaceName, Name: traitName}), gomock.Not(gomock.Nil()), gomock.Any()).
		Return(
			fmt.Errorf(serverErr),
		)

	reconciler := newLoggingTraitReconciler(mock, t)
	result, err := reconciler.Reconcile(context.TODO(), request)
	assert.Nil(err)
	assert.NotNil(result)
	assert.True(result.Requeue)

	mock = mocks.NewMockClient(mocker)
	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(types.NamespacedName{Namespace: namespaceName, Name: traitName}), gomock.Not(gomock.Nil()), gomock.Any()).
		Return(
			&errors.StatusError{ErrStatus: k8smeta.Status{Code: 404}},
		)

	reconciler = newLoggingTraitReconciler(mock, t)
	result, err = reconciler.Reconcile(context.TODO(), request)
	assert.Nil(err)
	assert.NotNil(result)
	assert.True(result.Requeue)
}

// TestCreateOrUpdateLoggingTraitNoWorkloadChild tests the creation/update/deletion of LoggingTrait when
// no child resources for workload exists
// GIVEN a LoggingTrait with workload reference
// WHEN there is no child resource of the workload
// THEN fall back to the original workload and complete the reconciliation
func TestReconcileTraitNoWorkloadChild(t *testing.T) {
	assert := asserts.New(t)
	mocker := gomock.NewController(t)
	mock := mocks.NewMockClient(mocker)

	trait := &vzapi.LoggingTrait{
		TypeMeta: k8smeta.TypeMeta{
			Kind: vzapi.LoggingTraitKind,
		},
		ObjectMeta: k8smeta.ObjectMeta{
			Name:      traitName,
			Namespace: namespaceName,
		},
		Spec: vzapi.LoggingTraitSpec{
			WorkloadReference: oamrt.TypedReference{
				APIVersion: oamcore.SchemeGroupVersion.Identifier(),
				Kind:       oamcore.ContainerizedWorkloadKind,
				Name:       workloadName,
				UID:        types.UID(workloadUID),
			}},
	}

	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(client.ObjectKey{Namespace: namespaceName, Name: workloadName}), gomock.Not(gomock.Nil()), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj *unstructured.Unstructured, opt ...client.GetOption) error {
			return nil
		})

	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(client.ObjectKey{Namespace: "", Name: workloadDefinitionNamespace}), gomock.Not(gomock.Nil()), gomock.Any()).
		Return(fmt.Errorf(serverErr))

	reconciler := newLoggingTraitReconciler(mock, t)
	result, supported, err := reconciler.reconcileTraitCreateOrUpdate(context.TODO(), vzlog.DefaultLogger(), trait)
	assert.NoError(err)
	assert.NotNil(result)
	assert.True(supported)

	mock = mocks.NewMockClient(mocker)
	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(client.ObjectKey{Namespace: namespaceName, Name: workloadName}), gomock.Not(gomock.Nil()), gomock.Any()).
		DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj *unstructured.Unstructured, opt ...client.GetOption) error {
			return nil
		})

	mock.EXPECT().
		Get(gomock.Any(), gomock.Eq(client.ObjectKey{Namespace: "", Name: workloadDefinitionNamespace}), gomock.Not(gomock.Nil()), gomock.Any()).
		Return(fmt.Errorf(serverErr))

	reconciler = newLoggingTraitReconciler(mock, t)
	result, err = reconciler.reconcileTraitDelete(context.TODO(), vzlog.DefaultLogger(), trait)
	assert.NoError(err)
	assert.NotNil(result)
}

// convertToUnstructured converts an object to an Unstructured version
// object - The object to convert to Unstructured
func convertToUnstructured(object interface{}) (unstructured.Unstructured, error) {
	jbytes, err := json.Marshal(object)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	var u map[string]interface{}
	_ = json.Unmarshal(jbytes, &u)
	return unstructured.Unstructured{Object: u}, nil
}

// appendAsUnstructured appends an object to the list after converting it to an Unstructured
// list - The list to append to.
// object - The object to convert to Unstructured and append to the list
func appendAsUnstructured(list *unstructured.UnstructuredList, object interface{}) error {
	u, err := convertToUnstructured(object)
	if err != nil {
		return err
	}
	list.Items = append(list.Items, u)
	return nil
}

// newLoggingTraitReconciler creates a new reconciler for testing
// cli - The Kerberos client to inject into the reconciler
func newLoggingTraitReconciler(cli client.Client, t *testing.T) LoggingTraitReconciler {
	scheme := runtime.NewScheme()
	vzapi.AddToScheme(scheme)
	reconciler := LoggingTraitReconciler{
		Client: cli,
		Log:    zap.S(),
		Scheme: scheme,
	}
	return reconciler
}

func newDeployment(deploymentName string, namespaceName string, workloadName string, workloadUID string) k8sapps.Deployment {
	return k8sapps.Deployment{
		TypeMeta: k8smeta.TypeMeta{
			APIVersion: k8sapps.SchemeGroupVersion.Identifier(),
			Kind:       "Deployment",
		},
		ObjectMeta: k8smeta.ObjectMeta{
			Name:              deploymentName,
			Namespace:         namespaceName,
			CreationTimestamp: k8smeta.Now(),
			OwnerReferences: []k8smeta.OwnerReference{
				{
					APIVersion: oamcore.SchemeGroupVersion.Identifier(),
					Kind:       oamcore.ContainerizedWorkloadKind,
					Name:       workloadName,
					UID:        types.UID(workloadUID),
				},
			},
		},
		Spec: k8sapps.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "test-container",
						},
					},
				},
			},
		},
	}
}

// TestReconcileKubeSystem tests to make sure we do not reconcile
// Any resource that belong to the kube-system namespace
func TestReconcileKubeSystem(t *testing.T) {
	assert := asserts.New(t)
	mocker := gomock.NewController(t)
	mock := mocks.NewMockClient(mocker)

	// create a request and reconcile it
	request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: vzconst.KubeSystem, Name: traitName}}
	reconciler := newLoggingTraitReconciler(mock, t)
	result, err := reconciler.Reconcile(context.TODO(), request)

	// Validate the results
	mocker.Finish()
	assert.Nil(err)
	assert.True(result.IsZero())
}
