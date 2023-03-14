#!/bin/bash
# Copyright (c) 2020, 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#While loop for verrazzano-application-operator to wait for webhooks to be started before starting up
./poll_webhook.sh "https://verrazzano-application-operator-webhook:443/validate-clusters-verrazzano-io-v1alpha1-multiclusterapplicationconfiguration"
./poll_webhook.sh "https://verrazzano-application-operator-webhook:443/validate-clusters-verrazzano-io-v1alpha1-multiclustercomponent"
./poll_webhook.sh "https://verrazzano-application-operator-webhook:443/validate-clusters-verrazzano-io-v1alpha1-multiclusterconfigmap"
./poll_webhook.sh "https://verrazzano-application-operator-webhook:443/validate-clusters-verrazzano-io-v1alpha1-multiclustersecret"
./poll_webhook.sh "https://verrazzano-application-operator-webhook:443/validate-clusters-verrazzano-io-v1alpha1-verrazzanoproject"
./poll_webhook.sh "https://verrazzano-application-operator-webhook:443/appconfig-defaulter"
./poll_webhook.sh "https://verrazzano-application-operator-webhook:443/validate-oam-verrazzano-io-v1alpha1-ingresstrait"
