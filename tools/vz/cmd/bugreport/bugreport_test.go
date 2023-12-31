// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package bugreport

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	vzconstants "github.com/verrazzano/verrazzano/pkg/constants"
	"github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1beta1"
	"github.com/verrazzano/verrazzano/tools/vz/pkg/constants"
	pkghelper "github.com/verrazzano/verrazzano/tools/vz/pkg/helpers"
	"github.com/verrazzano/verrazzano/tools/vz/test/helpers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testKubeConfig = "kubeconfig"
	testK8sContext = "testcontext"

// captureResourceErrMsg   = "Capturing resources from the cluster"
// sensitiveDataErrMsg     = "WARNING: Please examine the contents of the bug report for any sensitive data"
// captureLogErrMsg        = "Capturing log from pod verrazzano-platform-operator in verrazzano-install namespace"
// dummyNamespaceErrMsg    = "Namespace dummy not found in the cluster"
)

// TestBugReportHelp
// GIVEN a CLI bug-report command
// WHEN I call cmd.Help for bug-report
// THEN expect the help for the command in the standard output
func TestBugReportHelp(t *testing.T) {
	stdoutFile, stderrFile := createStdTempFiles(t)
	defer os.Remove(stdoutFile.Name())
	defer os.Remove(stderrFile.Name())

	rc := helpers.NewFakeRootCmdContext(genericclioptions.IOStreams{In: os.Stdin, Out: stdoutFile, ErrOut: stderrFile})
	cmd := NewCmdBugReport(rc)
	assert.NotNil(t, cmd)
	err := cmd.Help()
	if err != nil {
		assert.Error(t, err)
	}

	buf, err := os.ReadFile(stdoutFile.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(buf), "Verrazzano command line utility to collect data from the cluster, to report an issue")
}

// TestBugReportExistingReportFile
// GIVEN a CLI bug-report command using an existing file for flag --report-file
// WHEN I call cmd.Execute for bug-report
// THEN expect an error
func TestBugReportExistingReportFile(t *testing.T) {
	cmd := setUpandVerifyResources(t)

	tmpDir, _ := os.MkdirTemp("", "bug-report")
	defer cleanupTempDir(t, tmpDir)

	// Define and create the bug report file
	reportFile := "bug-report.tgz"
	bugRepFile, err := os.Create(tmpDir + string(os.PathSeparator) + reportFile)
	if err != nil {
		assert.Error(t, err)
	}
	defer cleanupFile(t, bugRepFile)

	setUpGlobalFlags(cmd)
	err = cmd.PersistentFlags().Set(constants.BugReportFileFlagName, bugRepFile.Name())
	assert.NoError(t, err)
	err = cmd.Execute()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "file exists")
}

// TestBugReportExistingDir
// GIVEN a CLI bug-report command with flag --report-file pointing to an existing directory
// WHEN I call cmd.Execute for bug-report
// THEN expect an error
func TestBugReportExistingDir(t *testing.T) {
	cmd := setUpandVerifyResources(t)

	tmpDir, _ := os.MkdirTemp("", "bug-report")
	defer cleanupTempDir(t, tmpDir)

	reportDir := tmpDir + string(os.PathSeparator) + "test-report"
	if err := os.Mkdir(reportDir, os.ModePerm); err != nil {
		assert.Error(t, err)
	}

	setUpGlobalFlags(cmd)
	err := cmd.PersistentFlags().Set(constants.BugReportFileFlagName, reportDir)
	assert.NoError(t, err)
	err = cmd.Execute()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "file exists")
}

// TestBugReportNonExistingFileDir
// GIVEN a CLI bug-report command with flag --report-file pointing to a file, where the directory doesn't exist
// WHEN I call cmd.Execute for bug-report
// THEN expect an error
func TestBugReportNonExistingFileDir(t *testing.T) {
	cmd := setUpandVerifyResources(t)

	tmpDir, _ := os.MkdirTemp("", "bug-report")
	defer cleanupTempDir(t, tmpDir)

	reportDir := tmpDir + string(os.PathSeparator) + "test-report"
	reportFile := reportDir + string(os.PathSeparator) + string(os.PathSeparator) + "bug-report.tgz"

	setUpGlobalFlags(cmd)
	err := cmd.PersistentFlags().Set(constants.BugReportFileFlagName, reportFile)
	assert.NoError(t, err)
	err = cmd.Execute()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

// TestBugReportFileNoPermission
// GIVEN a CLI bug-report command with flag --report-file pointing to a file, where there is no write permission
// WHEN I call cmd.Execute for bug-report
// THEN expect an error
func TestBugReportFileNoPermission(t *testing.T) {
	cmd := setUpandVerifyResources(t)

	tmpDir, _ := os.MkdirTemp("", "bug-report")
	defer cleanupTempDir(t, tmpDir)

	reportDir := tmpDir + string(os.PathSeparator) + "test-report"
	// Create a directory with only read permission
	if err := os.Mkdir(reportDir, 0400); err != nil {
		assert.Error(t, err)
	}
	reportFile := reportDir + string(os.PathSeparator) + "bug-report.tgz"
	setUpGlobalFlags(cmd)
	err := cmd.PersistentFlags().Set(constants.BugReportFileFlagName, reportFile)
	assert.NoError(t, err)
	err = cmd.Execute()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

// TestBugReportSuccess
// GIVEN a CLI bug-report command with multiple flags
// WHEN I call cmd.Execute
// THEN expect the command to show the resources captured in the standard output and create the bug report file
func TestBugReportSuccess(t *testing.T) {
	cmd := setUpandVerifyResources(t)

	tmpDir, _ := os.MkdirTemp("", "bug-report")
	defer cleanupTempDir(t, tmpDir)

	bugRepFile := tmpDir + string(os.PathSeparator) + "bug-report.tgz"
	setUpGlobalFlags(cmd)
	err := cmd.PersistentFlags().Set(constants.BugReportFileFlagName, bugRepFile)
	assert.NoError(t, err)
	err = cmd.PersistentFlags().Set(constants.BugReportIncludeNSFlagName, "dummy,verrazzano-install,default")
	assert.NoError(t, err)
	err = cmd.PersistentFlags().Set(constants.VerboseFlag, "true")
	assert.NoError(t, err)
	err = cmd.Execute()
	if err != nil {
		assert.Error(t, err)
	}
	assert.NoError(t, err)
}

// TestDefaultBugReportSuccess
// GIVEN a CLI bug-report command with no flags (default)
// WHEN I call cmd.Execute from user permissive directory
// THEN expect the command to show the resources captured in the standard output and create the bug report file in current dir
func TestDefaultBugReportSuccess(t *testing.T) {
	c := getClientWithVZWatch()

	// Verify the vz resource is as expected
	vz := v1beta1.Verrazzano{}
	err := c.Get(context.TODO(), types.NamespacedName{Namespace: "default", Name: "verrazzano"}, &vz)
	assert.NoError(t, err)

	stdoutFile, stderrFile := createStdTempFiles(t)
	defer os.Remove(stdoutFile.Name())
	defer os.Remove(stderrFile.Name())

	rc := helpers.NewFakeRootCmdContext(genericclioptions.IOStreams{In: os.Stdin, Out: stdoutFile, ErrOut: stderrFile})
	rc.SetClient(c)
	cmd := NewCmdBugReport(rc)
	assert.NotNil(t, cmd)
	setUpGlobalFlags(cmd)
	err = cmd.Execute()
	assert.Nil(t, err)

	if !pkghelper.CheckAndRemoveBugReportExistsInDir("") {
		t.Fatal("cannot find bug report file in current directory")
	}
}

// TestDefaultBugReportSuccess
// GIVEN a CLI bug-report command with no flags (default)
// WHEN I call cmd.Execute from read only directory
// THEN expect the command to show the resources captured in the standard output and create the bug report file in tmp dir
func TestDefaultBugReportReadOnlySuccess(t *testing.T) {
	c := getClientWithVZWatch()

	// Verify the vz resource is as expected
	vz := v1beta1.Verrazzano{}
	err := c.Get(context.TODO(), types.NamespacedName{Namespace: "default", Name: "verrazzano"}, &vz)
	assert.NoError(t, err)

	stdoutFile, stderrFile := createStdTempFiles(t)
	defer os.Remove(stdoutFile.Name())
	defer os.Remove(stderrFile.Name())

	rc := helpers.NewFakeRootCmdContext(genericclioptions.IOStreams{In: os.Stdin, Out: stdoutFile, ErrOut: stderrFile})
	rc.SetClient(c)
	cmd := NewCmdBugReport(rc)
	assert.NotNil(t, cmd)
	setUpGlobalFlags(cmd)
	pwd, err := os.Getwd()
	assert.Nil(t, err)
	assert.Nil(t, os.Chdir("/"))
	defer os.Chdir(pwd)

	err = cmd.Execute()
	assert.Nil(t, err)

	if !pkghelper.CheckAndRemoveBugReportExistsInDir(os.TempDir() + "/") {
		t.Fatal("cannot find bug report file in temp directory")
	}
}

// TestBugReportDefaultReportFile
// GIVEN a CLI bug-report command
// WHEN I call cmd.Execute, without specifying --report-file
// THEN expect the command to create the report vz-bug-report-*.tar.gz under the current directory
func TestBugReportDefaultReportFile(t *testing.T) {
	// clean up the bugreport file that is generated
	defer func(t *testing.T) {
		if !pkghelper.CheckAndRemoveBugReportExistsInDir("") {
			t.Fatal("cannot find and delete bug report file in current directory")
		}
	}(t)

	c := getClientWithVZWatch()

	// Verify the vz resource is as expected
	vz := v1beta1.Verrazzano{}
	err := c.Get(context.TODO(), types.NamespacedName{Namespace: "default", Name: "verrazzano"}, &vz)
	assert.NoError(t, err)

	stdoutFile, err := os.CreateTemp("", "tmpstdout")
	assert.NoError(t, err)
	defer os.Remove(stdoutFile.Name())

	stderrFile, err := os.CreateTemp("", "tmpstderr")
	assert.NoError(t, err)
	defer os.Remove(stderrFile.Name())

	rc := helpers.NewFakeRootCmdContext(genericclioptions.IOStreams{In: os.Stdin, Out: stdoutFile, ErrOut: stderrFile})
	rc.SetClient(c)
	cmd := NewCmdBugReport(rc)
	err = cmd.PersistentFlags().Set(constants.VerboseFlag, "true")
	setUpGlobalFlags(cmd)
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
	err = cmd.Execute()
	assert.NoError(t, err)

	_, err = os.ReadFile(stdoutFile.Name())
	assert.NoError(t, err)
}

// TestBugReportNoVerrazzano
// GIVEN a CLI bug-report command
// WHEN I call cmd.Execute without Verrazzano installed
// THEN expect the command to generate bug report
func TestBugReportNoVerrazzano(t *testing.T) {
	c := getClientWithWatch()
	stdoutFile, stderrFile := createStdTempFiles(t)
	defer os.Remove(stdoutFile.Name())
	defer os.Remove(stderrFile.Name())

	rc := helpers.NewFakeRootCmdContext(genericclioptions.IOStreams{In: os.Stdin, Out: stdoutFile, ErrOut: stderrFile})
	rc.SetClient(c)
	cmd := NewCmdBugReport(rc)
	assert.NotNil(t, cmd)

	tmpDir, _ := os.MkdirTemp("", "bug-report")
	defer cleanupTempDir(t, tmpDir)

	bugRepFile := tmpDir + string(os.PathSeparator) + "bug-report.tgz"
	setUpGlobalFlags(cmd)
	err := cmd.PersistentFlags().Set(constants.BugReportFileFlagName, bugRepFile)
	assert.NoError(t, err)
	err = cmd.PersistentFlags().Set(constants.BugReportIncludeNSFlagName, "dummy,verrazzano-install")
	assert.NoError(t, err)
	err = cmd.Execute()
	if err != nil {
		assert.Error(t, err)
	}

	errBuf, err := os.ReadFile(stderrFile.Name())
	assert.NoError(t, err)
	assert.NotContains(t, string(errBuf), "Verrazzano is not installed")
	assert.FileExists(t, bugRepFile)
}

// TestBugReportFailureUsingInvalidClient
// GIVEN a CLI bug-report command
// WHEN I call cmd.Execute without Verrazzano installed and using an invalid client
// THEN expect the command to fail with a message indicating Verrazzano is not installed and no resource captured
func TestBugReportFailureUsingInvalidClient(t *testing.T) {
	c := getInvalidClient()
	stdoutFile, stderrFile := createStdTempFiles(t)
	defer os.Remove(stdoutFile.Name())
	defer os.Remove(stderrFile.Name())

	rc := helpers.NewFakeRootCmdContext(genericclioptions.IOStreams{In: os.Stdin, Out: stdoutFile, ErrOut: stderrFile})
	rc.SetClient(c)
	cmd := NewCmdBugReport(rc)
	assert.NotNil(t, cmd)

	tmpDir, _ := os.MkdirTemp("", "bug-report")
	defer cleanupTempDir(t, tmpDir)

	bugRepFile := tmpDir + string(os.PathSeparator) + "bug-report.tgz"
	setUpGlobalFlags(cmd)
	err := cmd.PersistentFlags().Set(constants.BugReportFileFlagName, bugRepFile)
	assert.NoError(t, err)
	err = cmd.PersistentFlags().Set(constants.BugReportIncludeNSFlagName, "dummy,verrazzano-install")
	assert.NoError(t, err)
	err = cmd.Execute()
	if err != nil {
		assert.Error(t, err)
	}

	errBuf, err := os.ReadFile(stderrFile.Name())
	assert.NoError(t, err)
	assert.NotContains(t, string(errBuf), "Verrazzano is not installed")
	assert.FileExists(t, bugRepFile)
}

// getClientWithWatch returns a client containing all VPO objects
func getClientWithWatch() client.WithWatch {
	return fake.NewClientBuilder().WithScheme(pkghelper.NewScheme()).WithObjects(getVpoObjects()[1:]...).Build()
}

// getClientWithVZWatch returns a client containing all VPO objects and the Verrazzano CR
func getClientWithVZWatch() client.WithWatch {
	return fake.NewClientBuilder().WithScheme(pkghelper.NewScheme()).WithObjects(getVpoObjects()...).Build()
}

func getVpoObjects() []client.Object {
	return []client.Object{
		&v1beta1.Verrazzano{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "verrazzano",
			},
			Spec: v1beta1.VerrazzanoSpec{
				Profile: v1beta1.Dev,
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: vzconstants.VerrazzanoInstallNamespace,
				Name:      constants.VerrazzanoPlatformOperator,
				Labels: map[string]string{
					"app":               constants.VerrazzanoPlatformOperator,
					"pod-template-hash": "45f78ffddd",
				},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: vzconstants.VerrazzanoInstallNamespace,
				Name:      constants.VerrazzanoPlatformOperator,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": constants.VerrazzanoPlatformOperator},
				},
			},
			Status: appsv1.DeploymentStatus{
				AvailableReplicas: 1,
				UpdatedReplicas:   1,
			},
		},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: vzconstants.VerrazzanoInstallNamespace,
				Name:      fmt.Sprintf("%s-45f78ffddd", constants.VerrazzanoPlatformOperator),
				Annotations: map[string]string{
					"deployment.kubernetes.io/revision": "1",
				},
			},
		},
	}
}

// getInvalidClient returns an invalid client
func getInvalidClient() client.WithWatch {
	testObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "testnamespace",
			Name:      "testpod",
			Labels: map[string]string{
				"app":               "test-app",
				"pod-template-hash": "56f78ddcfd",
			},
		},
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "testnamespace",
			Name:      "testpod",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test-app"},
			},
		},
	}
	return fake.NewClientBuilder().WithScheme(pkghelper.NewScheme()).WithObjects(testObj, deployment).Build()
}

// cleanupTempDir cleans up the given temp directory after the test run
func cleanupTempDir(t *testing.T, dirName string) {
	if err := os.RemoveAll(dirName); err != nil {
		t.Fatalf("Remove directory failed: %v", err)
	}
}

// cleanupTempDir cleans up the given temp file after the test run
func cleanupFile(t *testing.T, file *os.File) {
	if err := file.Close(); err != nil {
		t.Fatalf("Close file failed: %v", err)
	}
	if err := os.Remove(file.Name()); err != nil {
		t.Fatalf("Close file failed: %v", err)
	}
}

// createStdTempFiles creates temporary files for stdout and stderr.
func createStdTempFiles(t *testing.T) (*os.File, *os.File) {
	stdoutFile, err := os.CreateTemp("", "tmpstdout")
	assert.NoError(t, err)

	stderrFile, err := os.CreateTemp("", "tmpstderr")
	assert.NoError(t, err)

	return stdoutFile, stderrFile
}

// TestBugReportSuccess
// GIVEN a CLI bug-report command
// WHEN I call cmd.Execute with include logs of  additional namespace and duration
// THEN expect the command to show the resources captured in the standard output and create the bug report file
func TestBugReportSuccessWithDuration(t *testing.T) {
	cmd := setUpandVerifyResources(t)

	tmpDir, _ := os.MkdirTemp("", "bug-report")
	defer cleanupTempDir(t, tmpDir)

	bugRepFile := tmpDir + string(os.PathSeparator) + "bug-report.tgz"
	setUpGlobalFlags(cmd)
	err := cmd.PersistentFlags().Set(constants.BugReportFileFlagName, bugRepFile)
	assert.NoError(t, err)
	err = cmd.PersistentFlags().Set(constants.BugReportIncludeNSFlagName, "dummy,verrazzano-install,default,test")
	assert.NoError(t, err)
	err = cmd.PersistentFlags().Set(constants.VerboseFlag, "true")
	assert.NoError(t, err)
	err = cmd.PersistentFlags().Set(constants.BugReportLogFlagName, "true")
	assert.NoError(t, err)
	// If invalid time value is given then error is expected
	err = cmd.PersistentFlags().Set(constants.BugReportTimeFlagName, "3t")
	assert.Error(t, err)
	// Valid time value
	err = cmd.PersistentFlags().Set(constants.BugReportTimeFlagName, "3h")
	assert.NoError(t, err)
	err = cmd.Execute()
	if err != nil {
		assert.Error(t, err)
	}
	assert.NoError(t, err)
}

func setUpGlobalFlags(cmd *cobra.Command) {
	tempKubeConfigPath, _ := os.CreateTemp(os.TempDir(), testKubeConfig)
	cmd.Flags().String(constants.GlobalFlagKubeConfig, tempKubeConfigPath.Name(), "")
	cmd.Flags().String(constants.GlobalFlagContext, testK8sContext, "")
}

func setUpandVerifyResources(t *testing.T) *cobra.Command {
	c := getClientWithVZWatch()

	// Verify the vz resource is as expected
	vz := v1beta1.Verrazzano{}
	err := c.Get(context.TODO(), types.NamespacedName{Namespace: "default", Name: "verrazzano"}, &vz)
	assert.NoError(t, err)

	stdoutFile, stderrFile := createStdTempFiles(t)
	defer os.Remove(stdoutFile.Name())
	defer os.Remove(stderrFile.Name())

	rc := helpers.NewFakeRootCmdContext(genericclioptions.IOStreams{In: os.Stdin, Out: stdoutFile, ErrOut: stderrFile})
	rc.SetClient(c)
	cmd := NewCmdBugReport(rc)
	assert.NotNil(t, cmd)

	return cmd
}
