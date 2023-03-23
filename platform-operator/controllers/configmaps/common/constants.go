// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package common

const (
	ConfigmapControllerLabel = "experimental.verrazzano.io/configmap-controller"
	ConfigmapAPIVersionLabel = "experimental.verrazzano.io/configmap-apiVersion"
	ConfigmapKindLabel       = "experimental.verrazzano.io/configmap-kind"
	ConfigmapGroupLabel      = "experimental.verrazzano.io/configmap-group"

	// ConfigMapObjectField The location for the delegate reconciler object
	ConfigMapObjectField = "object"
)
