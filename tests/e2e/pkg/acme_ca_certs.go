// Copyright (c) 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package pkg

import (
	"fmt"
	"net/http"
)

const letsEncryptStagingIntR3 = "https://letsencrypt.org/certs/staging/letsencrypt-stg-int-r3.pem"
const letsEncryptStagingIntE1 = "https://letsencrypt.org/certs/staging/letsencrypt-stg-int-e1.pem"

func getACMEStagingCAs() [][]byte {
	letsEncryptStagingIntE1CA := loadStagingCA(newSimpleHTTPClient(), letsEncryptStagingIntE1, "E1")
	letsEncryptStagingIntR3CA := loadStagingCA(newSimpleHTTPClient(), letsEncryptStagingIntR3, "R3")
	return [][]byte{letsEncryptStagingIntE1CA, letsEncryptStagingIntR3CA}
}

func newSimpleHTTPClient() *http.Client {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	httpClient := &http.Client{Transport: tr}
	return httpClient
}

func loadStagingCA(httpClient *http.Client, resURL string, caCertName string) []byte {
	resp, err := doReq(resURL, "GET", "", "", "", "", nil, newRetryableHTTPClient(httpClient))
	if err != nil {
		Log(Error, fmt.Sprintf("Error loading ACME staging CA: %v", err))
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		Log(Error, fmt.Sprintf("Unable to load ACME %s staging CA, status: %v\n", caCertName, resp.StatusCode))
		return nil
	}
	return resp.Body
}
