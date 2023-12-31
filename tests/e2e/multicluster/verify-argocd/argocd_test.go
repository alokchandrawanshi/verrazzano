// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package argocd_test

import (
	b64 "encoding/base64"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/verrazzano/verrazzano/pkg/k8s/resource"
	"github.com/verrazzano/verrazzano/pkg/k8sutil"
	"github.com/verrazzano/verrazzano/tests/e2e/multicluster/examples"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg"
	dump "github.com/verrazzano/verrazzano/tests/e2e/pkg/test/clusterdump"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/test/framework"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/test/framework/metrics"
)

const (
	waitTimeout          = 15 * time.Minute
	pollingInterval      = 10 * time.Second
	consistentlyDuration = 1 * time.Minute
	testNamespace        = "hello-helidon-argo"
	argoCDNamespace      = "argocd"
)

const (
	argoCDHelidonApplicationFile = "tests/e2e/multicluster/verify-argocd/testdata/hello-helidon-argocd-mc.yaml"
)

var expectedPodsHelloHelidon = []string{"helidon-config-deployment"}
var managedClusterName = os.Getenv("MANAGED_CLUSTER_NAME")
var adminKubeconfig = os.Getenv("ADMIN_KUBECONFIG")
var managedKubeconfig = os.Getenv("MANAGED_KUBECONFIG")

var t = framework.NewTestFramework("argocd_test")

var beforeSuite = t.BeforeSuiteFunc(func() {
	// Get the Hello Helidon Argo CD application yaml file
	// Deploy the Argo CD application in the admin cluster
	// This should internally deploy the helidon app to the managed cluster
	start := time.Now()
	Eventually(func() error {
		file, err := pkg.FindTestDataFile(argoCDHelidonApplicationFile)
		if err != nil {
			return err
		}
		return resource.CreateOrUpdateResourceFromFileInGeneratedNamespace(file, argoCDNamespace)
	}, waitTimeout, pollingInterval).ShouldNot(HaveOccurred(), "Failed to create Argo CD Application Project file")
	metrics.Emit(t.Metrics.With("deployment_elapsed_time", time.Since(start).Milliseconds()))

	beforeSuitePassed = true
})

var _ = BeforeSuite(beforeSuite)

var failed = false
var beforeSuitePassed = false

var _ = t.AfterEach(func() {
	failed = failed || CurrentSpecReport().Failed()
})

var _ = t.Describe("Multi Cluster Argo CD Validation", Label("f:platform-lcm.install"), func() {
	t.Context("Admin Cluster", func() {
		t.BeforeEach(func() {
			os.Setenv(k8sutil.EnvVarTestKubeConfig, os.Getenv("ADMIN_KUBECONFIG"))
		})

		t.It("has the expected secrets", func() {
			secretName := fmt.Sprintf("%s-argocd-cluster-secret", managedClusterName)
			Eventually(func() error {
				result, err := findSecret(argoCDNamespace, secretName)
				if result != false {
					pkg.Log(pkg.Error, fmt.Sprintf("Failed to get secret %s in namespace %s with error: %v", secretName, argoCDNamespace, err))
					return err
				}
				return err
			}, waitTimeout, pollingInterval).ShouldNot(HaveOccurred(), "Expected to find secret "+secretName)
		})

		t.It("secret has the data content with the same name as managed cluster", func() {
			secretName := fmt.Sprintf("%s-argocd-cluster-secret", managedClusterName)
			Eventually(func() error {
				result, err := findServerName(argoCDNamespace, secretName)
				if result != false {
					pkg.Log(pkg.Error, fmt.Sprintf("Failed to get servername in secret %s with error: %v", secretName, err))
					return err
				}
				return err
			}, waitTimeout, pollingInterval).ShouldNot(HaveOccurred(), "Expected to find managed cluster name "+managedClusterName)
		})
	})

	t.Context("Managed Cluster", func() {
		t.BeforeEach(func() {
			os.Setenv(k8sutil.EnvVarTestKubeConfig, os.Getenv("MANAGED_KUBECONFIG"))
		})
		// GIVEN an admin cluster and at least one managed cluster
		// WHEN the  example application has been  placed in managed cluster
		// THEN expect that the app is deployed to the managed cluster
		t.It("Has application placed", func() {
			Eventually(func() bool {
				result, err := helloHelidonPodsRunning(managedKubeconfig, testNamespace)
				if err != nil {
					pkg.Log(pkg.Error, fmt.Sprintf("One or more pods are not running in the namespace: %v, error: %v", testNamespace, err))
					return false
				}
				return result
			}, waitTimeout, pollingInterval).Should(BeTrue())
		})
	})

	t.Context("Delete resources", func() {
		t.BeforeEach(func() {
			os.Setenv(k8sutil.EnvVarTestKubeConfig, os.Getenv("ADMIN_KUBECONFIG"))
		})
		t.It("Delete resources on admin cluster", func() {
			Eventually(func() error {
				return deleteArgoCDApplication(adminKubeconfig)
			}, waitTimeout, pollingInterval).ShouldNot(HaveOccurred())
		})

		t.It("Verify automatic deletion on managed cluster", func() {
			Eventually(func() bool {
				return examples.VerifyAppDeleted(managedKubeconfig, testNamespace)
			}, consistentlyDuration, pollingInterval).Should(BeTrue())
		})

	})

})

var afterSuite = t.AfterSuiteFunc(func() {
	if failed || !beforeSuitePassed {
		dump.ExecuteBugReport(testNamespace)
	}
})

var _ = AfterSuite(afterSuite)

func deleteArgoCDApplication(kubeconfigPath string) error {
	start := time.Now()
	file, err := pkg.FindTestDataFile(argoCDHelidonApplicationFile)
	if err != nil {
		return err
	}
	if err := resource.DeleteResourceFromFileInCluster(file, kubeconfigPath); err != nil {
		return fmt.Errorf("failed to delete Argo CD hello-helidon application: %v", err)
	}

	metrics.Emit(t.Metrics.With("undeployment_elapsed_time", time.Since(start).Milliseconds()))
	return nil
}

func findSecret(namespace, name string) (bool, error) {
	s, err := pkg.GetSecret(namespace, name)
	if err != nil {
		return false, err
	}
	return s != nil, nil
}

func helloHelidonPodsRunning(kubeconfigPath string, namespace string) (bool, error) {
	result, err := pkg.PodsRunningInCluster(namespace, expectedPodsHelloHelidon, kubeconfigPath)
	if err != nil {
		pkg.Log(pkg.Error, fmt.Sprintf("One or more pods are not running in the namespace: %v, error: %v", namespace, err))
		return false, err
	}
	return result, nil
}

func findServerName(namespace, name string) (bool, error) {
	s, err := pkg.GetSecret(namespace, name)
	if err != nil {
		pkg.Log(pkg.Error, fmt.Sprintf("Failed to get secret %s in namespace %s with error: %v", name, namespace, err))
		return false, err
	}
	servername := string(s.Data["name"])
	decodeServerName, err := b64.StdEncoding.DecodeString(servername)
	if err != nil {
		pkg.Log(pkg.Error, fmt.Sprintf("Failed to decode secret data %s in secret %s with error: %v", servername, name, err))
		return false, err
	}
	return string(decodeServerName) != managedClusterName, nil
}
