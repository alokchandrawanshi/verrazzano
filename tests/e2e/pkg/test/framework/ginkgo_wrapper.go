// Copyright (c) 2021, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package framework

import (
	"fmt"
	"github.com/onsi/gomega"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/test"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/test/framework/metrics"
	"go.uber.org/zap"
	"os"
	"reflect"

	"github.com/onsi/ginkgo/v2"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg"
)

type TestFramework struct {
	Pkg     string
	Metrics *zap.SugaredLogger
	Logs    *zap.SugaredLogger
}

func NewTestFramework(pkg string) *TestFramework {
	t := new(TestFramework)

	t.Pkg = pkg

	logLevel, isSet := os.LookupEnv("FRAMEWORK_LOG_LEVEL")
	if !isSet {
		logLevel = "info"
	}
	t.Metrics, _ = metrics.NewLogger(pkg, metrics.MetricsIndex, logLevel)
	t.Logs, _ = metrics.NewLogger(pkg, metrics.TestLogIndex, logLevel, "stdout")

	t.initDumpDirectoryIfNecessary()
	return t
}

func (t *TestFramework) RegisterFailHandler() {
	gomega.RegisterFailHandler(t.Fail)
}

// initDumpDirectoryIfNecessary - sets the DUMP_DIRECTORY env variable to a default if not set externally
func (t *TestFramework) initDumpDirectoryIfNecessary() {
	if _, dumpDirIsSet := os.LookupEnv(test.DumpDirectoryEnvVarName); !dumpDirIsSet {
		dumpDirectory := t.Pkg
		dumpRoot, exists := os.LookupEnv(test.DumpRootDirectoryEnvVarName)
		if exists {
			dumpDirectory = fmt.Sprintf("%s/%s", dumpRoot, t.Pkg)
		}
		t.Logs.Infof("Defaulting %s to %s", test.DumpDirectoryEnvVarName, dumpDirectory)
		os.Setenv(test.DumpDirectoryEnvVarName, dumpDirectory)
	}
}

// AfterEach wraps Ginkgo AfterEach to emit a metric
func (t *TestFramework) AfterEach(args ...interface{}) bool {
	body := getFunctionBody(args...)
	f := func() {
		metrics.Emit(t.Metrics.With(metrics.Duration, metrics.DurationMillis()))
		reflect.ValueOf(body).Call([]reflect.Value{})
	}
	args[0] = f

	return ginkgo.AfterEach(args...)
}

// BeforeEach wraps Ginkgo BeforeEach to emit a metric
func (t *TestFramework) BeforeEach(args ...interface{}) bool {
	body := getFunctionBody(args...)
	f := func() {
		reflect.ValueOf(body).Call([]reflect.Value{})
	}
	args[0] = f
	return ginkgo.BeforeEach(args...)
}

// It wraps Ginkgo It to emit a metric
func (t *TestFramework) It(text string, args ...interface{}) bool {
	return ginkgo.It(text, t.MakeGinkgoArgs(args...)...)
}

func (t *TestFramework) WhenMeetsConditionFunc(condition string, conditionFunc func() (bool, error)) func(string, ...interface{}) bool {
	return func(text string, args ...interface{}) bool {
		met, err := conditionFunc()
		if err != nil {
			t.Logs.Errorf("Error checking condition %s: %v", condition, err)
			return false
		}
		if !met {
			t.Logs.Infof("Skipping test because condition is not met: %s", condition)
			return true
		}
		return ginkgo.It(text, t.MakeGinkgoArgs(args...)...)
	}
}

func (t *TestFramework) ItMinimumVersion(text string, version string, kubeconfigPath string, args ...interface{}) bool {
	supported, err := pkg.IsVerrazzanoMinVersion(version, kubeconfigPath)
	if err != nil {
		t.Logs.Errorf("Error getting Verrazzano version: %v", err)
		return false
	}
	if !supported {
		t.Logs.Infof("Skipping test because Verrazzano version is less than %s", version)
		return true
	}
	return ginkgo.It(text, t.MakeGinkgoArgs(args...)...)
}

func (t *TestFramework) MakeGinkgoArgs(args ...interface{}) []interface{} {
	body := getFunctionBodyPos(len(args)-1, args...)
	f := func() {
		metrics.Emit(t.Metrics) // Starting point metric
		reflect.ValueOf(body).Call([]reflect.Value{})
	}

	args[len(args)-1] = ginkgo.Offset(1)
	args = append(args, f)
	return args
}

// Describe wraps Ginkgo Describe to emit a metric
func (t *TestFramework) Describe(text string, args ...interface{}) bool {
	body := getFunctionBodyPos(len(args)-1, args...)
	f := func() {
		metrics.Emit(t.Metrics)
		reflect.ValueOf(body).Call([]reflect.Value{})
		metrics.Emit(t.Metrics.With(metrics.Duration, metrics.DurationMillis()))
	}
	args[len(args)-1] = ginkgo.Offset(1)
	args = append(args, f)
	return ginkgo.Describe(text, args...)
}

// DescribeTable - wrapper function for Ginkgo DescribeTable
func (t *TestFramework) DescribeTable(text string, args ...interface{}) bool {
	body := getFunctionBody(args...)
	funcType := reflect.TypeOf(body)
	f := reflect.MakeFunc(funcType, func(args []reflect.Value) (results []reflect.Value) {
		metrics.Emit(t.Metrics)
		rv := reflect.ValueOf(body).Call(args)
		metrics.Emit(t.Metrics.With(metrics.Duration, metrics.DurationMillis()))
		return rv
	})
	args[0] = f.Interface()
	return ginkgo.DescribeTable(text, args...)
}

// BeforeSuiteFunc wrap a function to be called with ginkgo.BeforeSuiteFunc. ginkgo.BeforeSuiteFunc
// // hard codes the call stack location, which requires calling it from the package level.
func (t *TestFramework) BeforeSuiteFunc(body func()) func() {
	t.failIfNilBody(body)
	f := func() {
		metrics.Emit(t.Metrics)
		reflect.ValueOf(body).Call([]reflect.Value{})
	}
	return f
}

// AfterSuiteFunc wrap a function to be called with ginkgo.AfterSuiteFunc. ginkgo.AfterSuiteFunc
// hard codes the call stack location, which requires calling it from the package level.
func (t *TestFramework) AfterSuiteFunc(body func()) func() {
	t.failIfNilBody(body)
	f := func() {
		metrics.Emit(t.Metrics.With(metrics.Duration, metrics.DurationMillis()))
		reflect.ValueOf(body).Call([]reflect.Value{})
	}
	return f
}

// Entry - wrapper function for Ginkgo Entry
func (t *TestFramework) Entry(description interface{}, args ...interface{}) ginkgo.TableEntry {
	// insert an Offset into the args, but not as the last item, so that the right code location is reported
	f := args[len(args)-1]
	args[len(args)-1] = ginkgo.Offset(6) // need to go 6 up the stack to get the caller
	args = append(args, f)
	return ginkgo.Entry(description, args...)
}

// Fail - wrapper function for Ginkgo Fail
func (t *TestFramework) Fail(message string, callerSkip ...int) {
	// Recover only to emit a fail, then re-panic
	defer func() {
		if p := recover(); p != nil {
			metrics.EmitFail(t.Metrics)
			panic(p)
		}
	}()
	ginkgo.Fail(message, callerSkip...)
}

// Context - wrapper function for Ginkgo Context
func (t *TestFramework) Context(text string, args ...interface{}) bool {
	return t.Describe(text, args...)
}

// When - wrapper function for Ginkgo When
func (t *TestFramework) When(text string, args ...interface{}) bool {
	return ginkgo.When(text, args...)
}

// SynchronizedBeforeSuite - wrapper function for Ginkgo SynchronizedBeforeSuite
func (t *TestFramework) SynchronizedBeforeSuite(process1Body func() []byte, allProcessBody func([]byte)) bool {
	return ginkgo.SynchronizedBeforeSuite(process1Body, allProcessBody)
}

// SynchronizedAfterSuite - wrapper function for Ginkgo SynchronizedAfterSuite
func (t *TestFramework) SynchronizedAfterSuite(allProcessBody func(), process1Body func()) bool {
	return ginkgo.SynchronizedAfterSuite(allProcessBody, process1Body)
}

// JustBeforeEach - wrapper function for Ginkgo JustBeforeEach
func (t *TestFramework) JustBeforeEach(args ...interface{}) bool {
	return ginkgo.JustBeforeEach(args...)
}

// JustAfterEach - wrapper function for Ginkgo JustAfterEach
func (t *TestFramework) JustAfterEach(args ...interface{}) bool {
	return ginkgo.JustAfterEach(args...)
}

// BeforeAll - wrapper function for Ginkgo BeforeAll
func (t *TestFramework) BeforeAll(args ...interface{}) bool {
	return ginkgo.BeforeAll(args...)
}

// AfterAll - wrapper function for Ginkgo AfterAll
func (t *TestFramework) AfterAll(args ...interface{}) bool {
	return ginkgo.AfterAll(args...)
}

// VzCurrentGinkgoTestDescription - wrapper function for ginkgo CurrentGinkgoTestDescription
func VzCurrentGinkgoTestDescription() ginkgo.SpecReport {
	pkg.Log(pkg.Debug, "VzCurrentGinkgoTestDescription wrapper")
	return ginkgo.CurrentSpecReport()
}

func (t *TestFramework) failIfNilBody(body func()) {
	if body == nil {
		ginkgo.Fail("Unsupported body type - expected non-nil")
	}
}

func failIfNotFunction(body interface{}) {
	if !isBodyFunc(body) {
		ginkgo.Fail("Unsupported body type - expected function")
	}
}

func getFunctionBodyPos(pos int, args ...interface{}) interface{} {
	if args == nil {
		ginkgo.Fail("Unsupported args type - expected non-nil")
	}
	body := args[pos]
	failIfNotFunction(body)
	return body
}

func getFunctionBody(args ...interface{}) interface{} {
	return getFunctionBodyPos(0, args...)
}
