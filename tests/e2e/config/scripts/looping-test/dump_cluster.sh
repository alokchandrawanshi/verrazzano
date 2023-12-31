#!/bin/bash -x

# Copyright (c) 2020, 2021, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

SCRIPT_DIR=$(cd $(dirname "$0"); pwd -P)
INCLUDE_CRDS=true

if [ -z "$1" ] ; then
  echo "Please provide directory to place resource dump"
  exit 1
fi

if [ -n "$2" ] ; then
  INCLUDE_CRDS=$2
fi

mkdir -p "$1"
cd "$1"
touch default.txt kube-node-lease.txt kube-public.txt kube-system.txt
cd "$SCRIPT_DIR"

echo "retrieving default resources"
"${SCRIPT_DIR}"/get_resources.sh default "${INCLUDE_CRDS}" > "$1"/default.txt

echo "retrieving kube-node-lease resources"
"${SCRIPT_DIR}"/get_resources.sh kube-node-lease "${INCLUDE_CRDS}" > "$1"/kube-node-lease.txt

echo "retrieving kube-public resources"
"${SCRIPT_DIR}"/get_resources.sh kube-public "${INCLUDE_CRDS}" > "$1"/kube-public.txt

echo "retrieving kube-system resources"
"${SCRIPT_DIR}"/get_resources.sh kube-system "${INCLUDE_CRDS}" > "$1"/kube-system.txt
