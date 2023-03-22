// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package components

import (
	"context"
	vzstring "github.com/verrazzano/verrazzano/pkg/string"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/yaml"
	"strconv"
	"time"

	vzctrl "github.com/verrazzano/verrazzano/pkg/controller"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	ConfigmapControllerAnnotation = "experimental.verrazzano.io/configmap-controller"
	ConfigmapAPIVersionAnnotation = "experimental.verrazzano.io/configmap-apiVersion"
	ConfigmapKindAnnotation       = "experimental.verrazzano.io/configmap-kind"
	ConfigmapGroupAnnotation      = "experimental.verrazzano.io/configmap-group"
)

type ConfigMapDelegateReconciler interface {
	// NewObject Returns a new instance of the object a reconciler manages
	NewObject() interface{}
	// Matches Allow delegate to indicate if it manages the specified GVK
	Matches(group string, version string, kind string) bool
	// DoReconcile reconciles the delegate object
	DoReconcile(log vzlog.VerrazzanoLogger, object interface{}) (ctrl.Result, error)
	DoDelete(log vzlog.VerrazzanoLogger, object interface{}) (ctrl.Result, error)
	FinalizerName() string
}

type ConfigMapReconciler struct {
	client.Client
	Name                    string
	Scheme                  *runtime.Scheme
	DelegateReconciler      ConfigMapDelegateReconciler
	NumConcurrentReconciles int
	DryRun                  bool
}

func (r *ConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	numConcurrentReconciles := r.NumConcurrentReconciles
	if numConcurrentReconciles <= 0 {
		numConcurrentReconciles = 1
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ConfigMap{}).
		WithEventFilter(r.createPredicate()).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: numConcurrentReconciles,
		}).
		Complete(r)
}

func (r *ConfigMapReconciler) createPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return r.objectMatches(e.Object)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return r.objectMatches(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return r.objectMatches(e.ObjectNew)
		},
	}
}

func (r *ConfigMapReconciler) isConfigmapControllerResource(cm *v1.ConfigMap) bool {
	cmController, found := cm.Annotations[ConfigmapControllerAnnotation]
	if !found {
		return false
	}
	isCMControllerObj, _ := strconv.ParseBool(cmController)
	return isCMControllerObj
}

func (r *ConfigMapReconciler) objectMatches(o client.Object) bool {
	configMap := o.(*v1.ConfigMap)

	if !r.isConfigmapControllerResource(configMap) {
		return false
	}

	if r.DelegateReconciler == nil {
		return false
	}

	// Check Delegate GVK
	var group, version, kind string
	var found bool
	if group, found = configMap.Annotations[ConfigmapGroupAnnotation]; !found {
		return false
	}
	if version, found = configMap.Annotations[ConfigmapAPIVersionAnnotation]; !found {
		return false
	}
	if kind, found = configMap.Annotations[ConfigmapKindAnnotation]; !found {
		return false
	}

	return r.DelegateReconciler.Matches(group, version, kind)
}

// Reconcile function for the ConfigMapReconciler
func (r *ConfigMapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	zap.S().Infof("Reconciling component configmap %s/%s", req.Namespace, req.Name)
	// Get the configmap for the request
	cm := v1.ConfigMap{}
	if err := r.Get(ctx, req.NamespacedName, &cm); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		zap.S().Errorf("Failed to get configmap %s/%s", req.Namespace, req.Name)
		return vzctrl.NewRequeueWithDelay(2, 3, time.Second), err
	}

	// Get the resource logger needed to log message using 'progress' and 'once' methods
	log, err := vzlog.EnsureResourceLogger(&vzlog.ResourceConfig{
		Name:           cm.Name,
		Namespace:      cm.Namespace,
		ID:             string(cm.UID),
		Generation:     cm.Generation,
		ControllerName: r.Name,
	})
	if err != nil {
		zap.S().Errorf("Failed to create controller logger for component configmap controller: %v", err)
		return vzctrl.NewRequeueWithDelay(2, 3, time.Second), err
	}

	if r.DelegateReconciler == nil {
		return ctrl.Result{}, log.ErrorfThrottledNewErr("No delegate reconciler defined for %s/%s", cm.Namespace, cm.Name)
	}

	if !r.isConfigmapControllerResource(&cm) {
		return vzctrl.NewRequeueWithDelay(5, 15, time.Second), log.ErrorfThrottledNewErr("Configmap %s/%s does not match this controller", cm.Namespace, cm.Name)
	}

	objectToReconcile := r.DelegateReconciler.NewObject()

	objDataYAML, objDataFound := cm.Data["object"]
	if !objDataFound {
		return vzctrl.NewRequeueWithDelay(5, 15, time.Second), log.ErrorfThrottledNewErr("No object data found for configmap %s/%s", cm.Namespace, cm.Name)
	}

	if err := yaml.Unmarshal([]byte(objDataYAML), objectToReconcile); err != nil {
		return vzctrl.NewRequeueWithDelay(5, 15, time.Second), log.ErrorfThrottledNewErr("Unable to unmarshal delegate object for configmap %s/%s, data: %s", cm.Namespace, cm.Name, objDataYAML)
	}

	// Check if resource is being deleted
	finalizerName := r.DelegateReconciler.FinalizerName()
	if !cm.ObjectMeta.DeletionTimestamp.IsZero() {
		// call delegate for deletion work and remove finalizer if no requeue is needed
		if result, err := r.DelegateReconciler.DoDelete(log, objectToReconcile); err != nil || !result.IsZero() {
			return result, err
		}
		log.Oncef("Removing finalizer %s", finalizerName)
		cm.ObjectMeta.Finalizers = vzstring.RemoveStringFromSlice(cm.ObjectMeta.Finalizers, finalizerName)
		if err := r.Update(ctx, &cm); err != nil {
			return newRequeueWithDelay(), err
		}
		return ctrl.Result{}, nil
	}

	if !vzstring.SliceContainsString(cm.ObjectMeta.Finalizers, finalizerName) {
		log.Debugf("Adding finalizer %s", finalizerName)
		cm.ObjectMeta.Finalizers = append(cm.ObjectMeta.Finalizers, finalizerName)
		if err := r.Update(context.TODO(), &cm); err != nil {
			return newRequeueWithDelay(), err
		}
	}

	result, err := r.DelegateReconciler.DoReconcile(log, objectToReconcile)
	if err != nil {
		return result, err
	}

	bytes, err := yaml.Marshal(objectToReconcile)
	if err != nil {
		return ctrl.Result{}, err
	}
	cm.Data["object"] = string(bytes)

	// Write any updates to the ConfigMap
	if err := r.Client.Update(context.TODO(), &cm); err != nil {
		return vzctrl.NewRequeueWithDelay(5, 15, time.Second), log.ErrorfThrottledNewErr("error updating configmap %s/%s for object, updated data: %s", cm.Namespace, cm.Name, objDataYAML)
	}

	log.Infof("Successfully reconciled %s/%s", cm.Namespace, cm.Name)
	return ctrl.Result{}, nil
}

func newRequeueWithDelay() ctrl.Result {
	return vzctrl.NewRequeueWithDelay(2, 5, time.Second)
}
