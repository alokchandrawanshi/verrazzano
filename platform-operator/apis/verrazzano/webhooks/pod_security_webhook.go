// Copyright (c) 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package webhooks

import (
	"context"
	"encoding/base64"
	vzlog "github.com/verrazzano/verrazzano/pkg/log"
	"github.com/verrazzano/verrazzano/tools/vz/pkg/constants"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/yaml"
	"strings"
)

var (
	podSecurityConfigMapName = "pod-security"
	ignoredNamespacesKey     = "ignored-namespaces"
	ignoredPodsKey           = "ignored-pods"
)

type podData struct {
	namespace string
	name      string
}

// PodSecurityWebhook type for mutating pods to add security settings
type PodSecurityWebhook struct {
	client.Client
	Decoder       *admission.Decoder
	KubeClient    kubernetes.Interface
	DynamicClient dynamic.Interface
	Defaulters    []MySQLDefaulter
}

type PodSecurityDefaulter interface {
}

// Handle is the entry point for the mutating webhook.
// This function is called for any jobs that are created in a namespace with the label istio-injection=enabled.
func (m *PodSecurityWebhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	var log = zap.S().With(vzlog.FieldResourceNamespace, req.Namespace, vzlog.FieldResourceName, req.Name, vzlog.FieldWebhook, "pod-security")
	pod := &corev1.Pod{}
	err := m.Decoder.Decode(req, pod)
	if err != nil {
		log.Error("Unable to decode pod due to ", zap.Error(err))
		return admission.Errored(http.StatusBadRequest, err)
	}
	skip, err := dontMutateSecurityForPod(m.Client, log, pod)
	if err != nil {
		log.Errorf("dont mutate: %v", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if skip {
		log.Info("-------- skipped ----------")

		return admission.Allowed("No action required, pod was designated to be ignored")
	}
	return mutatePod(req, pod, log)
}

// processJob processes the job request and applies the necessary annotations based on Job ownership and labels
func mutatePod(req admission.Request, pod *corev1.Pod, log *zap.SugaredLogger) admission.Response {
	var mutated bool
	for _, container := range pod.Spec.Containers {
		if mutateContainer(&container) {
			mutated = true
		}
	}
	for _, initContainer := range pod.Spec.InitContainers {
		if mutateContainer(&initContainer) {
			mutated = true
		}
	}
	if !mutated {
		log.Info("-------- not mutated ----------")

		admission.Allowed("No action required, pod already had pod security configured")
	}
	marshaledPodData, err := yaml.Marshal(pod)
	if err != nil {
		log.Error("Unable to marshall data for pod %s due to ", pod.Name, zap.Error(err))
		return admission.Errored(http.StatusInternalServerError, err)
	}
	log.Info("-------- marshalled ----------")
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPodData)
}

// InjectDecoder injects the decoder.
func (m *PodSecurityWebhook) InjectDecoder(d *admission.Decoder) error {
	m.Decoder = d
	return nil
}

func dontMutateSecurityForPod(c client.Client, log *zap.SugaredLogger, pod *corev1.Pod) (bool, error) {
	var cm corev1.ConfigMap

	if err := c.Get(context.TODO(), types.NamespacedName{
		Name:      podSecurityConfigMapName,
		Namespace: constants.VerrazzanoInstall,
	}, &cm); err != nil {
		log.Errorf("Error getting configmap %s: %v", podSecurityConfigMapName, err)
		return false, err
	}
	log.Info("-------- get ----------")

	ignoredNS := cm.Data[ignoredNamespacesKey]
	if ignoredNS != "" {
		decoded, err := base64.StdEncoding.DecodeString(ignoredNS)
		if err != nil {
			log.Errorf("Failed to decode configmap %s/%s data at key %s: %v", cm.Namespace, cm.Name, ignoredNamespacesKey, err)
			return false, err
		}

		var namespaces []string
		if err = yaml.Unmarshal(decoded, &namespaces); err != nil {
			log.Errorf("Failed to unmarshal namespaces : %v", err)
			return false, err
		}

		for _, ns := range namespaces {
			if pod.Namespace == ns {
				log.Debugf("Skipping update of security configuration for pod %s, pod is in ignored namespace %s", pod.Name, ns)
				return true, nil
			}
		}
	}
	log.Info("-------- namespace ----------")

	ignoredPods := cm.Data[ignoredPodsKey]
	if ignoredPods != "" {
		decoded, err := base64.StdEncoding.DecodeString(ignoredPods)
		if err != nil {
			log.Errorf("Failed to decode configmap %s/%s data at key %s: %v", cm.Namespace, cm.Name, ignoredPodsKey, err)
			return false, err
		}

		var pods []podData
		if err = yaml.Unmarshal(decoded, &pods); err != nil {
			log.Errorf("Failed to unmarshal pods : %v", err)
			return false, err
		}

		for _, ignoredPod := range pods {
			if pod.Namespace == ignoredPod.namespace && strings.Contains(pod.Name, ignoredPod.name) {
				log.Debugf("Skipping update of security configuration for pod %s, pod specified to be ignored %s", pod.Name)
				return true, nil
			}
		}
	}
	log.Info("-------- pods ----------")

	return false, nil
}

func mutateContainer(container *corev1.Container) bool {
	if container.SecurityContext == nil {
		container.SecurityContext = &corev1.SecurityContext{}
	}
	if container.SecurityContext.Capabilities == nil {
		container.SecurityContext.Capabilities = &corev1.Capabilities{}
	}
	for _, capability := range container.SecurityContext.Capabilities.Drop {
		if capability == "ALL" {
			return false
		}
	}
	container.SecurityContext.Capabilities.Drop = append(container.SecurityContext.Capabilities.Drop, "ALL")
	return true
}
