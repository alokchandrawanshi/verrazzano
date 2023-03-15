#!/bin/bash
# Copyright (c) 2023, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
SECONDS=0
MAX_SECONDS=120
webhook_url="$1"
while [ $SECONDS -lt $MAX_SECONDS ]; do
    http_code=$(curl --insecure -s -o /dev/null -w '%{http_code}' '$webhook_url' -H 'Content-Type: application/json')
    echo "$1 returned HTTP '$http_code'."
    if [[ "$http_code" != "200" ]]; then
        curl --insecure -v -H 'Content-Type: application/json' $webhook_url
        echo "waiting for 5 seconds"
        sleep 5
    else
        exit 0
    fi
done
exit 1
