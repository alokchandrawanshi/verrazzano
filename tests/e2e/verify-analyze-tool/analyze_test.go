// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package analyze

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8util "github.com/verrazzano/verrazzano/pkg/k8sutil"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/test/framework"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"os/exec"
	"strings"
	"time"
)

var (
	waitTimeout     = 10 * time.Second
	pollingInterval = 10 * time.Second
)

var t = framework.NewTestFramework("VZ Tool Analyze")
var _ = BeforeSuite(beforeSuite)
var _ = t.AfterEach(func() {})

var beforeSuite = t.BeforeSuiteFunc(func() {
})

var _ = t.Describe("VZ Tools", Label("f:verify-analyze-tool"), func() {
	t.Context("Start Analysis", func() {
		out, err := RunVzAnalyze()
		if err != nil {
			Fail(err.Error())
		}
		out, err = InjectIssues()
		if err != nil {
			Fail(err.Error())
		}
		out, err = RunVzAnalyze()
		if err != nil {
			Fail(err.Error())
		}
		t.It("Doesn't Have Image Pull Back Off Issue", func() {
			Eventually(func() bool {
				return testIssues(out, "ImagePullBackOff")
			}, waitTimeout, pollingInterval).Should(BeTrue())
		})
		out, err = RevertIssues()
		if err != nil {
			Fail(err.Error())
		}

		c, err := k8util.GetKubernetesClientset()
		if err != nil {
			Fail(err.Error())
		}
		deploymentsClient := c.ExtensionsV1beta1().Deployments("verrazzano-system")
		patch := []byte(`{"spec":{"template":{"spec":{"containers":[{"image":"ghcr.io/oracle/coherence-operator:3.YY","name":"coherence-operator"}]}}}}`)
		res, err := deploymentsClient.Patch(context.TODO(), "coherence-operator", types.MergePatchType, patch, v1.PatchOptions{}, "")
		if err != nil {
			Fail(err.Error())
		}
		fmt.Println(res)
	})
})

func testIssues(out, issueType string) bool {
	return strings.Contains(out, issueType)
}

// add issues in deployment
func InjectIssues() (string, error) {
	cmd := exec.Command("kubectl", "apply", "-f", "deployment.yaml")
	out, err := cmd.Output()
	return string(out), err
}

// reverse wrong deployment
func RevertIssues() (string, error) {
	cmd := exec.Command("kubectl", "apply", "-f", "undo_deployment.yaml")
	out, err := cmd.Output()
	return string(out), err

} //run vz analyze tool
func RunVzAnalyze() (string, error) {
	cmd := exec.Command("vz", "analyze")
	out, err := cmd.Output()
	return string(out), err
}
