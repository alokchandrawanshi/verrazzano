// Copyright (c) 2021, 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package helidon

import (
	"flag"
	"testing"

	"github.com/onsi/ginkgo/v2"
)

var skipDeploy bool
var skipUndeploy bool
var namespace string
var skipVerify bool
var istioInjection string
var helloHelidonAppConfig string
var helloHelidonComponent string

func init() {
	flag.BoolVar(&skipDeploy, "skipDeploy", false, "skipDeploy skips the call to install the application")
	flag.BoolVar(&skipUndeploy, "skipUndeploy", false, "skipUndeploy skips the call to install the application")
	flag.StringVar(&namespace, "namespace", generatedNamespace, "namespace is the app namespace")
	flag.BoolVar(&skipVerify, "skipVerify", false, "skipVerify skips the post deployment app validations")
	flag.StringVar(&istioInjection, "istioInjection", "enabled", "istioInjection enables the injection of istio side cars")
	flag.StringVar(&helloHelidonAppConfig, "appconfig", "", "appconfig is the the path to the desired Application Configuration to use")
	flag.StringVar(&helloHelidonComponent, "component", "", "component is the the path to the desired Component to use")
}

func TestHelidonExample(test *testing.T) {
	t.RegisterFailHandler()
	ginkgo.RunSpecs(test, "Hello Helidon Suite")
}
