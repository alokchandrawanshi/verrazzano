#!/bin/bash
# Copyright (c) 2022, 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
# While loop for verrazzano-cluster-operator to wait for webhooks to be started before starting up
./poll_webhook.sh "https://verrazzano-cluster-operator-webhook.verrazzano-system.svc.cluster.local:443/validate-clusters-verrazzano-io-v1alpha1-verrazzanomanagedcluster"
