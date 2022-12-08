// Copyright (c) 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package vmc

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/verrazzano/verrazzano/pkg/httputil"
	"io"
	corev1 "k8s.io/api/core/v1"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	clusterapi "github.com/verrazzano/verrazzano/cluster-operator/apis/clusters/v1alpha1"
	"github.com/verrazzano/verrazzano/pkg/log/vzlog"
	"github.com/verrazzano/verrazzano/platform-operator/constants"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/argocd"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/common"
	k8net "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	argocdAdminSecret = "argocd-initial-admin-secret"
	argocdTLSSecret   = "tls-argo-ingress"

	clustersAPIPath = "/api/v1/clusters"
	sessionPath     = "/api/v1/session"
)

type ArgoCDConfig struct {
	Host           string
	BaseURL        string
	ApiAccessToken string
}

type TlsClientConfig struct {
	CaData   string `json:"caData"`
	Insecure bool   `json:"insecure"`
}

type Config struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	TlsClientConfig TlsClientConfig
}

type PostPayload struct {
	ClusterResources bool `json:"clusterResources"`
	Config           Config
	Name             string `json:"name"`
	Server           string `json:"server"`
}

func (r *VerrazzanoManagedClusterReconciler) isArgoCDEnabled() bool {
	vz, _ := r.getVerrazzanoResource()
	return vz.Status.VerrazzanoInstance.ArgoCDURL != nil
}

// registerManagedClusterWithArgoCD calls the ArgoCD api to register a managed cluster with ArgoCD
func (r *VerrazzanoManagedClusterReconciler) registerManagedClusterWithArgoCD(ctx context.Context, vmc *clusterapi.VerrazzanoManagedCluster) error {
	if vmc.Status.ArgoCDRegistration.Status == "" {
		msg := fmt.Sprintf("Waiting for Verrazzano-created VMC named %s to have a cluster id in the status before ArgoCD cluster registration", vmc.Name)
		r.log.Progressf(msg)
		r.updateArgoCDStatus(ctx, vmc, clusterapi.RegistrationPendingRancher, msg)
	} else if vmc.Status.ArgoCDRegistration.Status == clusterapi.RegistrationMCResourceCreationCompleted {
		var clusterID = vmc.Status.RancherRegistration.ClusterID
		vz, err := r.getVerrazzanoResource()
		if err != nil {
			msg := "Could not find Verrazzano resource"
			r.updateArgoCDStatus(ctx, vmc, clusterapi.MCRegistrationFailed, msg)
			return r.log.ErrorfNewErr("Unable to find Verrazzano resource on admin cluster: %v", err)
		}
		if vz.Status.VerrazzanoInstance == nil {
			msg := "No instance information found in Verrazzano resource status"
			r.updateArgoCDStatus(ctx, vmc, clusterapi.MCRegistrationFailed, msg)
			return r.log.ErrorfNewErr("Unable to find instance information in Verrazzano resource status")
		}
		var rancherURL = "https://" + *(vz.Status.VerrazzanoInstance.RancherURL) + k8sClustersPath + clusterID

		caCert, err := common.GetRootCA(r.Client)
		if err != nil {
			msg := "Failed to get ArgoCD TLS CA"
			r.updateArgoCDStatus(ctx, vmc, clusterapi.MCRegistrationFailed, msg)
			return r.log.ErrorfNewErr("Unable to get ArgoCD TLS CA")
		}
		secret, err := getArgoCDAdminSecret(r.Client)
		if err != nil {
			msg := "Failed to get ArgoCD admin secret"
			r.updateArgoCDStatus(ctx, vmc, clusterapi.MCRegistrationFailed, msg)
			return r.log.ErrorfNewErr("Unable to get ArgoCD admin secret")
		}

		ac, err := newArgoCDConfig(r.Client, r.log)
		if err != nil {
			msg := "Failed to create ArgoCD API client"
			r.updateArgoCDStatus(ctx, vmc, clusterapi.MCRegistrationFailed, msg)
			return r.log.ErrorfNewErr("Unable to connect to ArgoCD API on admin cluster: %v", err)
		}

		// If the managed cluster is registered, we should not attempt to do POST
		isRegistered, err := isManagedClusterAlreadyExist(ac, vmc.Name, r.log)
		if err != nil {
			msg := "Failed to call ArgoCD clusters GET API"
			r.updateArgoCDStatus(ctx, vmc, clusterapi.MCRegistrationFailed, msg)
			return r.log.ErrorfNewErr("Unable to call ArgoCD clusters GET API on admin cluster: %v", err)
		}

		if !isRegistered {
			err = r.argocdClusterAdd(ac, vmc.Name, caCert, secret, rancherURL, r.log)
			if err != nil {
				msg := "Failed to call ArgoCD clusters POST API"
				r.updateArgoCDStatus(ctx, vmc, clusterapi.MCRegistrationFailed, msg)
				return r.log.ErrorfNewErr("Unable to call ArgoCD clusters POST API on admin cluster: %v", err)
			}
		}
		// TODO: invoke PUT if cluster already exists and caData are different
	}
	return nil
}

// isManagedClusterAlreadyExist returns true if the managed cluster does exist
func isManagedClusterAlreadyExist(ac *ArgoCDConfig, clusterName string, log vzlog.VerrazzanoLogger) (bool, error) {
	reqURL := "https://" + ac.Host + clustersAPIPath
	headers := map[string]string{"Authorization": "Bearer " + ac.ApiAccessToken}

	response, responseBody, err := sendHTTPRequest(http.MethodGet, reqURL, headers, "", ac, log)

	if response != nil && response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("tried to get cluster from Rancher but failed, response code: %d", response.StatusCode)
	}

	if err != nil {
		return false, err
	}

	name, err := httputil.ExtractFieldFromResponseBodyOrReturnError(responseBody, "name", "unable to find cluster state in Rancher response")
	if err != nil {
		return false, err
	}
	if name == clusterName {
		return true, nil
	}
	return false, nil
}

// makeClusterPayload returns the payload for Rancher cluster creation, given a cluster name
func newClusterPayload(clusterName string, caCert []byte, secret string, rancherURL string) (string, error) {
	payload := &PostPayload{
		ClusterResources: true,
		Config: Config{
			Username:        "admin",
			Password:        secret,
			TlsClientConfig: TlsClientConfig{CaData: string(caCert), Insecure: false},
		},
		Name:   clusterName,
		Server: rancherURL,
	}
	data, err := yaml.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(data), nil
}

// argocdClusterAdd simulates a client get request through the Rancher proxy for secrets
func (r *VerrazzanoManagedClusterReconciler) argocdClusterAdd(ac *ArgoCDConfig, clusterName string, caCert []byte, secret string, rancherURL string, log vzlog.VerrazzanoLogger) error {
	log.Debugf("Call the ArgoCD clusters api to register the cluster")
	action := http.MethodPost

	payload, err := newClusterPayload(clusterName, caCert, secret, rancherURL)
	if err != nil {
		return err
	}

	reqURL := "https://" + ac.Host + clustersAPIPath
	headers := map[string]string{"Authorization": "Bearer " + ac.ApiAccessToken}

	response, responseBody, err := sendHTTPRequest(action, reqURL, headers, payload, ac, log)

	if err != nil {
		return err
	}

	err = httputil.ValidateResponseCode(response, http.StatusCreated)
	if err != nil {
		return err
	}

	// TODO: parse the response and validate
	_, err = gabs.ParseJSON([]byte(responseBody))
	if err != nil {
		return err
	}

	log.Oncef("Successfully registered managed cluster in ArgoCD with name: %s", clusterName)
	return nil
}

// getArgoCACert the root CA certificate from the argocd TLS secret. If the secret does not exist, we
// return a nil slice.
func getArgoCACert(rdr client.Reader) ([]byte, error) {
	secret := &corev1.Secret{}
	nsName := types.NamespacedName{
		Namespace: constants.ArgoCDNamespace,
		Name:      argocdTLSSecret}

	if err := rdr.Get(context.TODO(), nsName, secret); err != nil {
		return nil, err
	}
	return secret.Data[common.ArgoCDCACert], nil
}

// getArgoCACert the root CA certificate from the argocd TLS secret. If the secret does not exist, we
// return a nil slice.
func getArgoCDAdminSecret(rdr client.Reader) (string, error) {
	secret := &corev1.Secret{}
	nsName := types.NamespacedName{
		Namespace: constants.ArgoCDNamespace,
		Name:      argocdAdminSecret}

	if err := rdr.Get(context.TODO(), nsName, secret); err != nil {
		return "", err
	}
	return string(secret.Data["password"]), nil
}

// newArgoCDConfig returns a populated ArgoCDConfig struct that can be used to make calls to the clusters API
func newArgoCDConfig(rdr client.Reader, log vzlog.VerrazzanoLogger) (*ArgoCDConfig, error) {
	ac := &ArgoCDConfig{BaseURL: "https://" + nginxIngressHostName}

	log.Debug("Getting ArgoCD ingress host name")
	hostname, err := getArgoCDIngressHostname(rdr)
	if err != nil {
		log.Errorf("Failed to get ArgoCD ingress host name: %v", err)
		return nil, err
	}
	ac.Host = hostname

	log.Once("Getting admin token from ArgoCD")
	adminToken, err := getAdminTokenFromArgoCD(rdr, ac, log)
	if err != nil {
		log.ErrorfThrottled("Failed to get admin token from Rancher: %v", err)
		return nil, err
	}
	ac.ApiAccessToken = adminToken

	return ac, nil
}

// getAdminTokenFromArgoCD does a login with ArgoCD and returns the token from the response
func getAdminTokenFromArgoCD(rdr client.Reader, ac *ArgoCDConfig, log vzlog.VerrazzanoLogger) (string, error) {
	secret, err := getArgoCDAdminSecret(rdr)
	if err != nil {
		return "", err
	}

	action := http.MethodPost
	payload := `{"Username": "admin", "Password": "` + secret + `"}`
	reqURL := ac.BaseURL + sessionPath
	headers := map[string]string{"Content-Type": "application/json"}

	response, responseBody, err := sendHTTPRequest(action, reqURL, headers, payload, ac, log)
	if err != nil {
		return "", err
	}

	err = httputil.ValidateResponseCode(response, http.StatusCreated)
	if err != nil {
		return "", err
	}

	return httputil.ExtractFieldFromResponseBodyOrReturnError(responseBody, "token", "unable to find token in Rancher response")
}

// sendRequest builds an HTTP request, sends it, and returns the response
func sendHTTPRequest(action string, reqURL string, headers map[string]string, payload string,
	ac *ArgoCDConfig, log vzlog.VerrazzanoLogger) (*http.Response, string, error) {

	req, err := http.NewRequest(action, reqURL, strings.NewReader(payload))
	if err != nil {
		return nil, "", err
	}

	req.Header.Add("Accept", "*/*")

	for k := range headers {
		req.Header.Add(k, headers[k])
	}
	req.Header.Add("Host", ac.Host)
	req.Host = ac.Host

	return doHTTPRequest(req, ac, log)
}

// doRequest configures an HTTP transport (including TLS), sends an HTTP request with retries, and returns the response
func doHTTPRequest(req *http.Request, ac *ArgoCDConfig, log vzlog.VerrazzanoLogger) (*http.Response, string, error) {
	log.Debugf("Attempting HTTP request: %v", req)

	proxyURL := getProxyURL()
	var tlsConfig *tls.Config
	tlsConfig = &tls.Config{
		ServerName: ac.Host,
		MinVersion: tls.VersionTLS12,
	}
	tr := &http.Transport{
		TLSClientConfig:       tlsConfig,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	// if we have a proxy, then set it in the transport
	if proxyURL != "" {
		u := url.URL{}
		proxy, err := u.Parse(proxyURL)
		if err != nil {
			return nil, "", err
		}
		tr.Proxy = http.ProxyURL(proxy)
	}

	client := &http.Client{Transport: tr, Timeout: 30 * time.Second}
	var resp *http.Response
	var err error

	// resp.Body is consumed by the first try, and then no longer available (empty)
	// so we need to read the body and save it so we can use it in each retry
	buffer, _ := io.ReadAll(req.Body)

	common.Retry(defaultRetry, log, true, func() (bool, error) {
		// update the body with the saved data to prevent the "zero length body" error
		req.Body = io.NopCloser(bytes.NewBuffer(buffer))
		resp, err = rancherHTTPClient.Do(client, req)

		// check for a network error and retry
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			log.Infof("Temporary error executing HTTP request %v : %v, retrying", req, nerr)
			return false, err
		}

		// if err is another kind of network error that is not considered "temporary", then retry
		if err, ok := err.(*url.Error); ok {
			if err, ok := err.Err.(*net.OpError); ok {
				if derr, ok := err.Err.(*net.DNSError); ok {
					log.Infof("DNS error: %v, retrying", derr)
					return false, err
				}
			}
		}

		// retry any HTTP 500 errors
		if resp != nil && resp.StatusCode >= 500 && resp.StatusCode <= 599 {
			log.ErrorfThrottled("HTTP status %v executing HTTP request %v, retrying", resp.StatusCode, req)
			return false, err
		}

		// if err is some other kind of unexpected error, retry
		if err != nil {
			return false, err
		}
		return true, err
	})

	if err != nil {
		return resp, "", err
	}
	defer resp.Body.Close()

	// extract the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return resp, string(body), err
}

// getArgoCDIngressHostname gets the ArgoCD ingress host name. This is used to set the host for TLS.
func getArgoCDIngressHostname(rdr client.Reader) (string, error) {
	ingress := &k8net.Ingress{}
	nsName := types.NamespacedName{
		Namespace: argocd.ComponentNamespace,
		Name:      constants.ArgoCDIngress}
	if err := rdr.Get(context.TODO(), nsName, ingress); err != nil {
		return "", fmt.Errorf("Failed to get ArgoCD ingress %v: %v", nsName, err)
	}

	if len(ingress.Spec.Rules) > 0 {
		// the first host will do
		return ingress.Spec.Rules[0].Host, nil
	}

	return "", fmt.Errorf("Failed, ArgoCD ingress %v is missing host names", nsName)
}

// Update the ArgoCD registration status
func (r *VerrazzanoManagedClusterReconciler) updateArgoCDStatus(ctx context.Context, vmc *clusterapi.VerrazzanoManagedCluster, status clusterapi.ArgoCDRegistrationStatus, message string) {
	// Skip the update if the status has not changed
	if vmc.Status.ArgoCDRegistration.Status == status &&
		vmc.Status.ArgoCDRegistration.Message == message {
		return
	}
	vmc.Status.ArgoCDRegistration.Status = status
	vmc.Status.ArgoCDRegistration.Message = message

	// Fetch the existing VMC to avoid conflicts in the status update
	existingVMC := &clusterapi.VerrazzanoManagedCluster{}
	err := r.Get(context.TODO(), types.NamespacedName{Namespace: vmc.Namespace, Name: vmc.Name}, existingVMC)
	if err != nil {
		r.log.Errorf("Failed to get the existing VMC %s from the cluster: %v", vmc.Name, err)
	}
	existingVMC.Status.ArgoCDRegistration = vmc.Status.ArgoCDRegistration

	err = r.Status().Update(ctx, existingVMC)
	if err != nil {
		r.log.Errorf("Failed to update ArgoCD registration status for VMC %s: %v", vmc.Name, err)
	}
}
