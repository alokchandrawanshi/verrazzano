// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package scenario

import (
	"fmt"
	"github.com/verrazzano/verrazzano/tools/vz/pkg/helpers"

	helmcli "github.com/verrazzano/verrazzano/pkg/helm"
)

var UninstallFunc = helmcli.Uninstall

// StopScenarioByID stops a running scenario specified by the scenario ID
func (m ScenarioMananger) StopScenarioByID(ID string, vzHelper helpers.VZHelper) error {
	cm, err := m.getConfigMapByID(ID)
	if err != nil {
		return err
	}
	sc, err := m.getScenarioFromConfigmap(cm)
	if err != nil {
		return err
	}
	// Delete Helm releases
	for _, h := range sc.HelmReleases {
		if m.Verbose {
			fmt.Fprintf(vzHelper.GetOutputStream(), "Uninstalling Helm release %s/%s\n", h.Namespace, h.Name)
		}
		err := UninstallFunc(m.Log, h.Name, h.Namespace, m.DryRun)
		if err != nil {
			return err
		}
	}
	// Delete config map
	err = m.deleteConfigMap(cm)
	if err != nil {
		return err
	}
	return nil
}
