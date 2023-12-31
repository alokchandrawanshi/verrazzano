// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package k8sclient

import (
	"fmt"
	vzv1alpha1 "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	vpoClient "github.com/verrazzano/verrazzano/platform-operator/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	k8sapiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/dynamic"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

// PsrClient contains the client sets for accessing various resource
type PsrClient struct {
	CrtlRuntime client.Client
	VzInstall   vpoClient.Interface
	DynClient   dynamic.Interface
}

// NewPsrClient returns a PSR client.
// Try several times to get the client.  This fixes timing issue
// where connections fails if Istio sidecar not ready
func NewPsrClient() (PsrClient, error) {
	const max = 15
	var retErr error
	for i := 1; i <= max; i++ {
		c, err := tryNewPsrClient()
		if err == nil {
			return c, nil
		}
		retErr = err
		time.Sleep(1 * time.Second)
	}
	return PsrClient{}, retErr
}

// tryNewPsrClient returns the set of clients used by PSR
func tryNewPsrClient() (PsrClient, error) {
	p := PsrClient{}

	// Create the controller runtime client and add core resources to the scheme
	// Along with Verrazzano
	cfg, err := controllerruntime.GetConfig()
	if err != nil {
		return PsrClient{}, fmt.Errorf("Failed to get controller-runtime config %v", err)
	}
	p.CrtlRuntime, err = client.New(cfg, client.Options{})
	if err != nil {
		return PsrClient{}, fmt.Errorf("Failed to create a controller-runtime client %v", err)
	}
	_ = corev1.AddToScheme(p.CrtlRuntime.Scheme())
	_ = k8sapiext.AddToScheme(p.CrtlRuntime.Scheme())
	_ = vzv1alpha1.AddToScheme(p.CrtlRuntime.Scheme())

	// Create the client for accessing the Verrazzano API
	p.VzInstall, err = vpoClient.NewForConfig(cfg)
	if err != nil {
		return PsrClient{}, err
	}
	p.DynClient, err = dynamic.NewForConfig(cfg)
	if err != nil {
		return PsrClient{}, err
	}
	return p, nil
}
