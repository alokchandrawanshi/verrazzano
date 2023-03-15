#!/bin/bash
# Copyright (c) 2020, 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#While loop for verrazzano-application-operator to wait for webhooks to be started before starting up
SCRIPT_DIR=$(
    cd $(dirname "$0")
    pwd -P
)
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-application-operator-webhook:443/validate-clusters-verrazzano-io-v1alpha1-multiclusterapplicationconfiguration"
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-application-operator-webhook:443/validate-clusters-verrazzano-io-v1alpha1-multiclustercomponent"
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-application-operator-webhook:443/validate-clusters-verrazzano-io-v1alpha1-multiclusterconfigmap"
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-application-operator-webhook:443/validate-clusters-verrazzano-io-v1alpha1-multiclustersecret"
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-application-operator-webhook:443/validate-clusters-verrazzano-io-v1alpha1-verrazzanoproject"
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-application-operator-webhook:443/appconfig-defaulter"
${SCRIPT_DIR}/poll_webhook.sh "https://verrazzano-application-operator-webhook:443/validate-oam-verrazzano-io-v1alpha1-ingresstrait"
