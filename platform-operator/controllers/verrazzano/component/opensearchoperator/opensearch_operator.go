// Copyright (c) 2023, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package opensearchoperator

import (
	"context"
	"fmt"
	"github.com/verrazzano/verrazzano/pkg/bom"
	"github.com/verrazzano/verrazzano/pkg/k8s/ready"
	vzapi "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	installv1beta1 "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1beta1"
	"github.com/verrazzano/verrazzano/platform-operator/controllers/verrazzano/component/spi"
	"github.com/verrazzano/verrazzano/platform-operator/internal/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

const (
	opensearchOperatorDeploymentName = "opensearch-operator-controller-manager"

	opsterOSDIngressName = "opster-osd"

	opsterOSIngressName      = "opster-os"
	securityconfigSecretName = "securityconfig-secret"
)

const securityconfigSecret = `
apiVersion: v1
kind: Secret
metadata:
  name: securityconfig-secret
  namespace: verrazzano-logging
type: Opaque
stringData:
  action_groups.yml: |-
    _meta:
      type: "actiongroups"
      config_version: 2
  internal_users.yml: |-
    _meta:
      type: "internalusers"
      config_version: 2
    admin:
      hash: "$2y$12$lJsHWchewGVcGlYgE3js/O4bkTZynETyXChAITarCHLz8cuaueIyq"
      reserved: true
      backend_roles:
      - "admin"
      description: "Demo admin user"
    dashboarduser:
      hash: "$2a$12$4AcgAt3xwOWadA5s5blL6ev39OXDNhmOesEoo33eZtrq2N0YrU3H."
      reserved: true
      description: "Demo OpenSearch Dashboards user"
  nodes_dn.yml: |-
    _meta:
      type: "nodesdn"
      config_version: 2
  whitelist.yml: |-
    _meta:
      type: "whitelist"
      config_version: 2
  tenants.yml: |-
    _meta:
      type: "tenants"
      config_version: 2
  roles_mapping.yml: |-
    _meta:
      type: "rolesmapping"
      config_version: 2
    vz_log_pusher:
      reserved: false
      backend_roles:
      - "vz_log_pusher"
      description: "Maps to Verrazzano role which has access to push logs to verrazzano specific indices"
    all_access:
      reserved: false
      backend_roles:
      - "admin"
      description: "Maps admin to all_access"
      users:
      - "*"
    own_index:
      reserved: false
      users:
      - "*"
      description: "Allow full access to an index named like the username"
    readall:
      reserved: false
      backend_roles:
      - "readall"
    manage_snapshots:
      reserved: false
      backend_roles:
      - "snapshotrestore"
    dashboard_server:
      reserved: true
      users:
      - "dashboarduser"
  roles.yml: |-
    _meta:
      type: "roles"
      config_version: 2
    dashboard_read_only:
      reserved: true
    security_rest_api_access:
      reserved: true
    # Log pusher for verrazzano
    vz_log_pusher:
      reserved: false
      hidden: false
      cluster_permissions:
      - cluster_all
      index_permissions:
      - index_patterns:
        - "verrazzano-*"
        allowed_actions:
          - crud
    # Allows users to view monitors, destinations and alerts
    alerting_read_access:
      reserved: true
      cluster_permissions:
        - 'cluster:admin/opendistro/alerting/alerts/get'
        - 'cluster:admin/opendistro/alerting/destination/get'
        - 'cluster:admin/opendistro/alerting/monitor/get'
        - 'cluster:admin/opendistro/alerting/monitor/search'
    # Allows users to view and acknowledge alerts
    alerting_ack_alerts:
      reserved: true
      cluster_permissions:
        - 'cluster:admin/opendistro/alerting/alerts/*'
    # Allows users to use all alerting functionality
    alerting_full_access:
      reserved: true
      cluster_permissions:
        - 'cluster_monitor'
        - 'cluster:admin/opendistro/alerting/*'
      index_permissions:
        - index_patterns:
            - '*'
          allowed_actions:
            - 'indices_monitor'
            - 'indices:admin/aliases/get'
            - 'indices:admin/mappings/get'
    # Allow users to read Anomaly Detection detectors and results
    anomaly_read_access:
      reserved: true
      cluster_permissions:
        - 'cluster:admin/opendistro/ad/detector/info'
        - 'cluster:admin/opendistro/ad/detector/search'
        - 'cluster:admin/opendistro/ad/detectors/get'
        - 'cluster:admin/opendistro/ad/result/search'
        - 'cluster:admin/opendistro/ad/tasks/search'
        - 'cluster:admin/opendistro/ad/detector/validate'
        - 'cluster:admin/opendistro/ad/result/topAnomalies'
    # Allows users to use all Anomaly Detection functionality
    anomaly_full_access:
      reserved: true
      cluster_permissions:
        - 'cluster_monitor'
        - 'cluster:admin/opendistro/ad/*'
      index_permissions:
        - index_patterns:
            - '*'
          allowed_actions:
            - 'indices_monitor'
            - 'indices:admin/aliases/get'
            - 'indices:admin/mappings/get'
    # Allows users to read Notebooks
    notebooks_read_access:
      reserved: true
      cluster_permissions:
        - 'cluster:admin/opendistro/notebooks/list'
        - 'cluster:admin/opendistro/notebooks/get'
    # Allows users to all Notebooks functionality
    notebooks_full_access:
      reserved: true
      cluster_permissions:
        - 'cluster:admin/opendistro/notebooks/create'
        - 'cluster:admin/opendistro/notebooks/update'
        - 'cluster:admin/opendistro/notebooks/delete'
        - 'cluster:admin/opendistro/notebooks/get'
        - 'cluster:admin/opendistro/notebooks/list'
    # Allows users to read observability objects
    observability_read_access:
      reserved: true
      cluster_permissions:
        - 'cluster:admin/opensearch/observability/get'
    # Allows users to all Observability functionality
    observability_full_access:
      reserved: true
      cluster_permissions:
        - 'cluster:admin/opensearch/observability/create'
        - 'cluster:admin/opensearch/observability/update'
        - 'cluster:admin/opensearch/observability/delete'
        - 'cluster:admin/opensearch/observability/get'
    # Allows users to read and download Reports
    reports_instances_read_access:
      reserved: true
      cluster_permissions:
        - 'cluster:admin/opendistro/reports/instance/list'
        - 'cluster:admin/opendistro/reports/instance/get'
        - 'cluster:admin/opendistro/reports/menu/download'
    # Allows users to read and download Reports and Report-definitions
    reports_read_access:
      reserved: true
      cluster_permissions:
        - 'cluster:admin/opendistro/reports/definition/get'
        - 'cluster:admin/opendistro/reports/definition/list'
        - 'cluster:admin/opendistro/reports/instance/list'
        - 'cluster:admin/opendistro/reports/instance/get'
        - 'cluster:admin/opendistro/reports/menu/download'
    # Allows users to all Reports functionality
    reports_full_access:
      reserved: true
      cluster_permissions:
        - 'cluster:admin/opendistro/reports/definition/create'
        - 'cluster:admin/opendistro/reports/definition/update'
        - 'cluster:admin/opendistro/reports/definition/on_demand'
        - 'cluster:admin/opendistro/reports/definition/delete'
        - 'cluster:admin/opendistro/reports/definition/get'
        - 'cluster:admin/opendistro/reports/definition/list'
        - 'cluster:admin/opendistro/reports/instance/list'
        - 'cluster:admin/opendistro/reports/instance/get'
        - 'cluster:admin/opendistro/reports/menu/download'
    # Allows users to use all asynchronous-search functionality
    asynchronous_search_full_access:
      reserved: true
      cluster_permissions:
        - 'cluster:admin/opendistro/asynchronous_search/*'
      index_permissions:
        - index_patterns:
            - '*'
          allowed_actions:
            - 'indices:data/read/search*'
    # Allows users to read stored asynchronous-search results
    asynchronous_search_read_access:
      reserved: true
      cluster_permissions:
        - 'cluster:admin/opendistro/asynchronous_search/get'
    # Allows user to use all index_management actions - ism policies, rollups, transforms
    index_management_full_access:
      reserved: true
      cluster_permissions:
        - "cluster:admin/opendistro/ism/*"
        - "cluster:admin/opendistro/rollup/*"
        - "cluster:admin/opendistro/transform/*"
      index_permissions:
        - index_patterns:
            - '*'
          allowed_actions:
            - 'indices:admin/opensearch/ism/*'
    # Allows users to use all cross cluster replication functionality at leader cluster
    cross_cluster_replication_leader_full_access:
      reserved: true
      index_permissions:
        - index_patterns:
            - '*'
          allowed_actions:
            - "indices:admin/plugins/replication/index/setup/validate"
            - "indices:data/read/plugins/replication/changes"
            - "indices:data/read/plugins/replication/file_chunk"
    # Allows users to use all cross cluster replication functionality at follower cluster
    cross_cluster_replication_follower_full_access:
      reserved: true
      cluster_permissions:
        - "cluster:admin/plugins/replication/autofollow/update"
      index_permissions:
        - index_patterns:
            - '*'
          allowed_actions:
            - "indices:admin/plugins/replication/index/setup/validate"
            - "indices:data/write/plugins/replication/changes"
            - "indices:admin/plugins/replication/index/start"
            - "indices:admin/plugins/replication/index/pause"
            - "indices:admin/plugins/replication/index/resume"
            - "indices:admin/plugins/replication/index/stop"
            - "indices:admin/plugins/replication/index/update"
            - "indices:admin/plugins/replication/index/status_check"
  config.yml: |-
    _meta:
      type: "config"
      config_version: "2"
    config:
      dynamic:
        kibana:
          multitenancy_enabled: false
          server_username: kibanaserver
        do_not_fail_on_forbidden: true
        http:
          anonymous_auth_enabled: true
          xff:
            enabled: true
            internalProxies: '.*' # trust all internal proxies, regex pattern. need to put auth proxy ip in future, it should have the ip addresses of auth proxy services
            remoteIpHeader: 'x-forwarded-for'
        authc:
          proxy_auth_domain:
            description: "Authenticate via proxy"
            http_enabled: true
            transport_enabled: true
            order: 0
            http_authenticator:
              type: proxy
              challenge: false
              config:
                user_header: "X-WEBAUTH-USER"
                roles_header: "x-proxy-roles"
            authentication_backend:
              type: noop
    
          basic_internal_auth_domain:
            description: "Authenticate via HTTP Basic against internal users database"
            http_enabled: true
            transport_enabled: true
            order: 1
            http_authenticator:
              type: basic
              challenge: false
            authentication_backend:
              type: intern
`

func createSecurityconfigSecret(ctx spi.ComponentContext) error {
	client := ctx.Client()
	log := ctx.Log()

	secretName := types.NamespacedName{
		Namespace: ComponentNamespace,
		Name:      securityconfigSecret,
	}

	secret := &v1.Secret{}
	err := client.Get(context.TODO(), secretName, secret)

	// Create the secret if it doesn't exist
	if err != nil && errors.IsNotFound(err) {
		log.Debugf("%s not found. Creating", securityconfigSecret)
		secret := &v1.Secret{}
		if err := yaml.Unmarshal([]byte(securityconfigSecret), secret); err != nil {
			return err
		}
		if err := client.Create(context.TODO(), secret); err != nil {
			return log.ErrorfNewErr("Failed to create %s for opensearch cluster: %v", securityconfigSecret, err)
		}
		return nil
	}

	return err
}

// GetOverrides gets the install overrides
func GetOverrides(object runtime.Object) interface{} {
	if effectiveCR, ok := object.(*vzapi.Verrazzano); ok {
		if effectiveCR.Spec.Components.OpenSearchOperator != nil {
			return effectiveCR.Spec.Components.OpenSearchOperator.ValueOverrides
		}
		return []vzapi.Overrides{}
	} else if effectiveCR, ok := object.(*installv1beta1.Verrazzano); ok {
		if effectiveCR.Spec.Components.OpenSearchOperator != nil {
			return effectiveCR.Spec.Components.OpenSearchOperator.ValueOverrides
		}
		return []installv1beta1.Overrides{}
	}

	return []vzapi.Overrides{}
}

// AppendOverrides appends the default overrides for the component
func AppendOverrides(_ spi.ComponentContext, _ string, _ string, _ string, kvs []bom.KeyValue) ([]bom.KeyValue, error) {
	bomFile, err := bom.NewBom(config.GetDefaultBOMFilePath())
	if err != nil {
		return kvs, err
	}
	image, err := bomFile.BuildImageOverrides(ComponentName)
	if err != nil {
		return kvs, err
	}
	return append(kvs, image...), nil
}

func (o opensearchOperatorComponent) isReady(ctx spi.ComponentContext) bool {
	return ready.DeploymentsAreReady(ctx.Log(), ctx.Client(), getDeploymentList(), 1, getPrefix(ctx))
}

func getDeploymentList() []types.NamespacedName {
	return []types.NamespacedName{
		{
			Name:      opensearchOperatorDeploymentName,
			Namespace: ComponentNamespace,
		},
	}
}

func getIngressList() []types.NamespacedName {
	return []types.NamespacedName{
		{
			Name:      opsterOSIngressName,
			Namespace: ComponentNamespace,
		},
		{
			Name:      opsterOSDIngressName,
			Namespace: ComponentNamespace,
		},
	}
}

func getPrefix(ctx spi.ComponentContext) string {
	return fmt.Sprintf("Component %s", ctx.GetComponent())
}
