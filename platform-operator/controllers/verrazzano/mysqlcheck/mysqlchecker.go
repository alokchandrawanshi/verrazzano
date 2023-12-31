// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package mysqlcheck

import (
	"time"

	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"go.uber.org/zap"
	clipkg "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	controllerName    = "MySQLChecker"
	channelBufferSize = 100
)

// mySQLChecker - holds global instance of MySQLChecker.  Required by MySQL workaround
// functions that don't have access to the MySQLChecker context.
var mySQLChecker *MySQLChecker

// MySQLChecker periodically checks the state of MySQL related pods and repairs
// any that are found to be in a stuck state (e.g. terminating, waiting for readiness gates).
type MySQLChecker struct {
	client        clipkg.Client
	tickTime      time.Duration
	RepairTimeout time.Duration
	log           vzlog.VerrazzanoLogger
	shutdown      chan int // The channel on which shutdown signals are sent/received
}

// NewMySQLChecker - instantiate a MySQLChecker context
func NewMySQLChecker(c clipkg.Client, tick time.Duration, timeout time.Duration) (*MySQLChecker, error) {
	log, err := vzlog.EnsureResourceLogger(&vzlog.ResourceConfig{
		Name:           componentName,
		Namespace:      componentNamespace,
		ID:             controllerName,
		Generation:     0,
		ControllerName: controllerName,
	})
	if err != nil {
		zap.S().Errorf("Failed to create resource logger for %s: %v", controllerName, err)
		return nil, err
	}

	mySQLChecker = &MySQLChecker{
		client:        c,
		tickTime:      tick,
		RepairTimeout: timeout,
		log:           log,
	}
	return mySQLChecker, nil
}

// GetMySQLChecker - returns the value of mySQLChecker
func GetMySQLChecker() *MySQLChecker {
	return mySQLChecker
}

// Start starts the MySQLChecker if it is not already running.
// It is safe to call Start multiple times, additional goroutines will not be created
func (mc *MySQLChecker) Start() {
	if mc.shutdown != nil {
		// already running, so nothing to do
		return
	}
	mc.shutdown = make(chan int, channelBufferSize)

	// goroutine updates availability every p.tickTime. If a shutdown signal is received (or channel is closed),
	// the goroutine returns.
	go func() {
		var err error
		ticker := time.NewTicker(mc.tickTime)
		for {
			select {
			case <-ticker.C:
				// timer event causes MySQL checks
				if err = mc.RepairMySQLPodStuckDeleting(); err != nil {
					mc.log.ErrorfThrottled("Failed to repair MySQL pods stuck terminating: %v", err)
				}
				if err = mc.RepairMySQLPodsWaitingReadinessGates(); err != nil {
					mc.log.ErrorfThrottled("Failed to repair MySQL pods waiting for readiness gates: %v", err)
				}
				if err = mc.RepairMySQLRouterPodsCrashLoopBackoff(); err != nil {
					mc.log.ErrorfThrottled("Failed to repair mysql-router pods stuck in CrashLoopBackoff: %v", err)
				}
			case <-mc.shutdown:
				// shutdown event causes termination
				ticker.Stop()
				return
			}
		}
	}()
}

// Pause pauses the MySQLChecker if it was running.
// It is safe to call Pause multiple times
func (mc *MySQLChecker) Pause() {
	if mc.shutdown != nil {
		close(mc.shutdown)
		mc.shutdown = nil
	}
}
