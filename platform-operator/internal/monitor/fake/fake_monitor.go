// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package fake

import "github.com/verrazzano/verrazzano/platform-operator/internal/monitor"

// BackgroundProcessMonitorType - a fake monitor object.
type BackgroundProcessMonitorType struct {
	Result    bool
	Err       error
	Running   bool
	Completed bool
}

func (f *BackgroundProcessMonitorType) CheckResult() (bool, error)           { return f.Result, f.Err }
func (f *BackgroundProcessMonitorType) Reset()                               {}
func (f *BackgroundProcessMonitorType) IsRunning() bool                      { return f.Running }
func (f *BackgroundProcessMonitorType) IsCompleted() bool                    { return f.Completed }
func (f *BackgroundProcessMonitorType) SetCompleted()                        { f.Completed = true; f.Running = false }
func (f *BackgroundProcessMonitorType) Run(operation monitor.BackgroundFunc) {}

// Check that &BackgroundProcessMonitorType implements BackgroundProcessMonitor
var _ monitor.BackgroundProcessMonitor = &BackgroundProcessMonitorType{}
