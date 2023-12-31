// Copyright (c) 2022, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package certac

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	aocnst "github.com/verrazzano/verrazzano/application-operator/constants"
	"github.com/verrazzano/verrazzano/cluster-operator/controllers/vmc"
	"github.com/verrazzano/verrazzano/pkg/constants"
	"github.com/verrazzano/verrazzano/pkg/mcconstants"
	vzapi "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	pocnst "github.com/verrazzano/verrazzano/platform-operator/constants"
	"github.com/verrazzano/verrazzano/tests/e2e/multicluster"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/test/framework"
	"github.com/verrazzano/verrazzano/tests/e2e/pkg/update"
	"sigs.k8s.io/yaml"
)

var (
	t               = framework.NewTestFramework("update fluentd external opensearch")
	adminCluster    *multicluster.Cluster
	managedClusters []*multicluster.Cluster
	originalCM      *vzapi.CertManagerComponent
	originalFluentd *vzapi.FluentdComponent
	waitTimeout     = 10 * time.Minute
	pollingInterval = 5 * time.Second
)

type CertModifier struct {
	CertManager *vzapi.CertManagerComponent
}

func (u *CertModifier) ModifyCR(cr *vzapi.Verrazzano) {
	cr.Spec.Components.CertManager = u.CertManager
}

var beforeSuite = t.BeforeSuiteFunc(func() {
	adminCluster = multicluster.AdminCluster()
	managedClusters = multicluster.ManagedClusters()
	adminVZ := adminCluster.GetCR(true)
	originalCM = adminVZ.Spec.Components.CertManager
	if !isDefaultCM(originalCM) {
		Fail("TestAdminClusterCertManagerUpdate requires default CertManager in AdminCluster")
	}
	originalFluentd = adminVZ.Spec.Components.Fluentd
	verifyRegistration()
})

var _ = BeforeSuite(beforeSuite)

var afterSuite = t.AfterSuiteFunc(func() {
})

var _ = AfterSuite(afterSuite)

var _ = t.Describe("Update admin-cluster cert-manager", Label("f:platform-lcm.update"), func() {
	t.Describe("multicluster cert-manager verify", Label("f:platform-lcm.multicluster-verify"), func() {
		t.It("admin-cluster cert-manager custom CA", func() {
			start := time.Now()
			oldIngressCaCrt := updateAdminClusterCA()
			verifyCaSync(oldIngressCaCrt)
			// verify new logs are flowing after updating admin cert
			verifyManagedFluentd(start)
		})
	})
	t.Describe("multicluster cert-manager verify cleanup", Label("f:platform-lcm.multicluster-verify"), func() {
		t.It("admin-cluster cert-manager revert to default self-signed CA", func() {
			start := time.Now()
			oldIngressCaCrt := revertToDefaultCertManager()
			verifyCaSync(oldIngressCaCrt)
			verifyManagedFluentd(start)
		})
	})
})

func updateAdminClusterCA() string {
	oldIngressCaCrt := adminCluster.
		GetSecretDataAsString(constants.VerrazzanoSystemNamespace, pocnst.VerrazzanoIngressSecret, mcconstants.CaCrtKey)
	genCA := adminCluster.GenerateCA()
	newCM := &vzapi.CertManagerComponent{
		Certificate: vzapi.Certificate{CA: vzapi.CA{SecretName: genCA, ClusterResourceNamespace: constants.CertManagerNamespace}},
	}
	m := &CertModifier{CertManager: newCM}
	update.RetryUpdate(m, adminCluster.KubeConfigPath, true, pollingInterval, waitTimeout)
	return oldIngressCaCrt
}

func isDefaultCM(cm *vzapi.CertManagerComponent) bool {
	return cm == nil || reflect.DeepEqual(*cm, vzapi.CertManagerComponent{})
}

func isDefaultFluentd(fl *vzapi.FluentdComponent) bool {
	return fl == nil || reflect.DeepEqual(*fl, vzapi.FluentdComponent{})
}

func revertToDefaultCertManager() string {
	oldIngressCaCrt := adminCluster.
		GetSecretDataAsString(constants.VerrazzanoSystemNamespace, pocnst.VerrazzanoIngressSecret, mcconstants.CaCrtKey)
	m := &CertModifier{CertManager: originalCM}
	update.RetryUpdate(m, adminCluster.KubeConfigPath, true, pollingInterval, waitTimeout)
	return oldIngressCaCrt
}

func verifyCaSync(oldIngressCaCrt string) {
	newIngressCaCrt := verifyAdminClusterIngressCA(oldIngressCaCrt)
	reapplyManagedClusterRegManifest(newIngressCaCrt)
	verifyCaSyncToManagedClusters(newIngressCaCrt)
}

func verifyAdminClusterIngressCA(oldIngressCaCrt string) string {
	start := time.Now()
	gomega.Eventually(func() bool {
		newIngressCaCrt := adminCluster.
			GetSecretDataAsString(constants.VerrazzanoSystemNamespace, pocnst.VerrazzanoIngressSecret, mcconstants.CaCrtKey)
		if newIngressCaCrt == oldIngressCaCrt {
			pkg.Log(pkg.Error, fmt.Sprintf("%v of %v is not updated", pocnst.VerrazzanoIngressSecret, adminCluster.Name))
		} else {
			pkg.Log(pkg.Error, fmt.Sprintf("%v of %v took %v updated", pocnst.VerrazzanoIngressSecret, adminCluster.Name, time.Since(start)))
		}
		return newIngressCaCrt != oldIngressCaCrt
	}).WithTimeout(waitTimeout).WithPolling(pollingInterval).Should(gomega.BeTrue(), fmt.Sprintf("Fail updating ingress %v CA", adminCluster.Name))
	return adminCluster.
		GetSecretDataAsString(constants.VerrazzanoSystemNamespace, pocnst.VerrazzanoIngressSecret, mcconstants.CaCrtKey)
}

func verifyCaSyncToManagedClusters(admCaCrt string) {
	for _, managedCluster := range managedClusters {
		verifyManagedClusterRegistration(managedCluster, admCaCrt, mcconstants.AdminCaBundleKey)
		verifyManagedClusterAdminKubeconfig(managedCluster, admCaCrt, mcconstants.KubeconfigKey)
		if isDefaultFluentd(originalFluentd) {
			verifyManagedClusterRegistration(managedCluster, admCaCrt, mcconstants.ESCaBundleKey)
		}
	}
}

func verifyManagedClusterAdminKubeconfig(managedCluster *multicluster.Cluster, admCaCrt, kubeconfigKey string) {
	start := time.Now()
	gomega.Eventually(func() error {
		newKubeconfig := managedCluster.
			GetSecretDataAsString(constants.VerrazzanoSystemNamespace, aocnst.MCAgentSecret, kubeconfigKey)
		jsonKubeconfig, err := yaml.YAMLToJSON([]byte(newKubeconfig))
		if err != nil {
			fullErr := fmt.Errorf("Failed converting admin kubeconfig to JSON: %v", err)
			pkg.Log(pkg.Error, fullErr.Error())
			return fullErr
		}
		parsedKubeconfig, err := gabs.ParseJSON(jsonKubeconfig)
		if err != nil {
			return fmt.Errorf("failed parsing admin kubeconfig JSON: %v", err)
		}
		newCaCrtData := parsedKubeconfig.Search("clusters", "0", "cluster", "certificate-authority-data").Data()
		if newCaCrtData == nil {
			return fmt.Errorf("kubeconfig in %s of %s has nil clusters[0].cluster.certificate-authority-data", aocnst.MCAgentSecret, managedCluster.Name)
		}
		newCaCrt, err := base64.StdEncoding.DecodeString(newCaCrtData.(string))
		if err != nil {
			return fmt.Errorf("could not decode certificate-authority-data in the kubeconfig in %s: %v", aocnst.MCAgentSecret, err)
		}
		if admCaCrt == string(newCaCrt) {
			pkg.Log(pkg.Info, fmt.Sprintf("%v of %v took %v updated", aocnst.MCAgentSecret, managedCluster.Name, time.Since(start)))
			return nil
		}
		err = fmt.Errorf("%v of %v is not updated", aocnst.MCAgentSecret, managedCluster.Name)
		pkg.Log(pkg.Error, err.Error())
		return err
	}).WithTimeout(waitTimeout).WithPolling(pollingInterval).Should(gomega.Not(gomega.HaveOccurred()))
}

func verifyManagedClusterRegistration(managedCluster *multicluster.Cluster, admCaCrt, cakey string) {
	start := time.Now()
	gomega.Eventually(func() bool {
		newCaCrt := managedCluster.
			GetSecretDataAsString(constants.VerrazzanoSystemNamespace, aocnst.MCRegistrationSecret, cakey)
		if newCaCrt == admCaCrt {
			pkg.Log(pkg.Error, fmt.Sprintf("%v of %v took %v updated", aocnst.MCRegistrationSecret, managedCluster.Name, time.Since(start)))
		} else {
			pkg.Log(pkg.Error, fmt.Sprintf("%v of %v is not updated", aocnst.MCRegistrationSecret, managedCluster.Name))
		}
		return admCaCrt == newCaCrt
	}).WithTimeout(waitTimeout).WithPolling(pollingInterval).Should(gomega.BeTrue(), fmt.Sprintf("Sync CA %v", managedCluster.Name))
}

func verifyManagedFluentd(since time.Time) {
	for _, managedCluster := range managedClusters {
		gomega.Eventually(func() bool {
			logs := managedCluster.FluentdLogs(5, since)
			ok := checkFluentdLogs(logs)
			if !ok {
				pkg.Log(pkg.Error, fmt.Sprintf("%v Fluentd is not ready: \n%v\n", managedCluster.Name, logs))
			}
			return ok
		}).WithTimeout(waitTimeout).WithPolling(pollingInterval).Should(gomega.BeTrue(), fmt.Sprintf("scrape target of %s is not ready", managedCluster.Name))
	}
}

func checkFluentdLogs(logs string) bool {
	return !strings.Contains(strings.ToUpper(logs), "ERROR") && !strings.Contains(logs, "Exception")
}

func verifyRegistration() {
	for _, managedCluster := range managedClusters {
		reg, _ := adminCluster.GetRegistration(managedCluster.Name)
		if reg == nil {
			adminCluster.Register(managedCluster)
			gomega.Eventually(func() bool {
				reg, err := adminCluster.GetRegistration(managedCluster.Name)
				return reg != nil && err == nil
			}).WithTimeout(waitTimeout).WithPolling(pollingInterval).Should(gomega.BeTrue(), fmt.Sprintf("%s is not registered", managedCluster.Name))
		}
	}
}

// reapplyManagedClusterRegManifest reapplies the registration manifest on managed clusters. For
// self-signed certs, this manual step is expected of users so that the admin kubeconfig used by the
// managed clusters can be reliably updated to use the new CA cert. Otherwise intermittent timing
// related failures may occur
func reapplyManagedClusterRegManifest(newCACert string) {
	for _, managedCluster := range managedClusters {
		waitForManifestSecretUpdated(managedCluster.Name, newCACert)
		gomega.Eventually(func() error {
			reg, err := adminCluster.GetManifest(managedCluster.Name)
			if err != nil {
				pkg.Log(pkg.Error, fmt.Sprintf("could not get manifest for managed cluster %v error: %v", managedCluster.Name, err))
				return err
			}
			pkg.Log(pkg.Info, fmt.Sprintf("Reapplying registration manifest for managed cluster %s", managedCluster.Name))
			managedCluster.Apply(reg)
			return nil
		}).WithTimeout(waitTimeout).WithPolling(pollingInterval).Should(gomega.Not(gomega.HaveOccurred()), fmt.Sprintf("Reapply registration manifest failed for cluster %s", managedCluster.Name))
	}
}

func waitForManifestSecretUpdated(managedClusterName string, newCACert string) {
	start := time.Now()
	manifestSecretName := vmc.GetManifestSecretName(managedClusterName)
	gomega.Eventually(func() error {
		manifestBytes, err := adminCluster.GetManifest(managedClusterName)
		if err != nil {
			pkg.Log(pkg.Error, fmt.Sprintf("Could not get manifest secret for managed cluster %s: %v", managedClusterName, err))
			return err
		}
		resourceContainer, err := extractResourceFromManifestYAML(manifestBytes, aocnst.MCAgentSecret)
		if err != nil {
			return err
		}

		encodedKubeconfigData := resourceContainer.Search("data", mcconstants.KubeconfigKey)
		if encodedKubeconfigData == nil {
			return fmt.Errorf("could not find admin kubeconfig data in manifest")
		}
		updated := kubeconfigContainsCACert(encodedKubeconfigData, newCACert)
		if updated {
			pkg.Log(pkg.Info, fmt.Sprintf("%s took %v updated", manifestSecretName, time.Since(start)))
			return nil
		}
		pkg.Log(pkg.Info, fmt.Sprintf("%s not updated", manifestSecretName))
		return fmt.Errorf("manifest secret for cluster %s not updated with new CA cert", managedClusterName)
	}).WithTimeout(waitTimeout).WithPolling(pollingInterval).
		Should(gomega.Not(gomega.HaveOccurred()))
}

func extractResourceFromManifestYAML(manifestBytes []byte, resourceName string) (*gabs.Container, error) {
	yamlDocs := bytes.Split(manifestBytes, []byte("---\n"))
	for _, yamlDoc := range yamlDocs {
		jsonDoc, err := yaml.YAMLToJSON(yamlDoc)
		if err != nil {
			pkg.Log(pkg.Error, fmt.Sprintf("Could not parse YAML in manifest to JSON %v", err))
			return nil, err
		}
		resourceContainer, err := gabs.ParseJSON(jsonDoc)
		if err != nil {
			pkg.Log(pkg.Error, fmt.Sprintf("Could not parse JSON manifest secret: %v", err))
			return nil, err
		}
		if resourceContainer.Search("metadata", "name").Data() == resourceName {
			return resourceContainer, nil
		}
	}
	return nil, fmt.Errorf("resource %s not found in manifest", resourceName)
}

func kubeconfigContainsCACert(encodedKubeconfigData *gabs.Container, newCACert string) bool {
	kubeconfig, err := base64.StdEncoding.DecodeString(encodedKubeconfigData.Data().(string))
	if err != nil {
		pkg.Log(pkg.Error, fmt.Sprintf("Could not base64 decode kubeconfig in manifest: %v", err))
		return false
	}
	encodedCACert := base64.StdEncoding.EncodeToString([]byte(newCACert))
	return strings.Contains(string(kubeconfig), encodedCACert)
}
