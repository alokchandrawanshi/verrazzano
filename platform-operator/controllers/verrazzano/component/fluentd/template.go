// Copyright (c) 2022, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package fluentd

import (
	yaml2 "github.com/verrazzano/verrazzano/pkg/yaml"
	vzapi "github.com/verrazzano/verrazzano/platform-operator/apis/verrazzano/v1alpha1"
	"sigs.k8s.io/yaml"
)

const baseTemplate = `
    {
      "index_patterns":[
        "verrazzano-system",
        "verrazzano-application*"
      ],
      "version":60001,
      "priority": 101,
      "data_stream": {},
      "template": {
        "settings":{
          "index.refresh_interval":"5s",
          "index.mapping.total_fields.limit":"2000",
          "number_of_shards":5,
          "index.number_of_replicas":0,
          "index.auto_expand_replicas":"0-1"
        },
        "mappings":{
          "dynamic_templates":[
            {
              "message_field":{
                "path_match":"message",
                "match_mapping_type":"string",
                "mapping":{
                  "type":"text",
                  "norms":false
                }
              }
            },
            {
              "object_fields": {
                "match": "*",
                "match_mapping_type": "object",
                "mapping": {
                  "type": "object"
                }
              }
            },
            {
              "all_non_object_fields":{
                "match":"*",
                "mapping":{
                  "type":"text",
                  "norms":false,
                  "fields":{
                    "keyword":{
                      "type":"keyword",
                      "ignore_above":256
                    }
                  }
                }
              }
            }
          ],
          "properties" : {
            "@timestamp": { "type": "date", "format": "strict_date_time||strict_date_optional_time||epoch_millis"}
          }
        }
      }
    }
`

func mergeIndexTemplates(vz *vzapi.Verrazzano) (string, error) {
	baseYaml, err := yaml.JSONToYAML([]byte(baseTemplate))
	if err != nil {
		return "", err
	}
	customTemplates := vz.Spec.Components.Fluentd.IndexTemplates

	customTemplateYamls := []string{}

	for _, template := range customTemplates {
		templateJSON, err := yaml.Marshal(template.Template)
		if err != nil {
			return "", err
		}

		templateYaml, err := yaml.JSONToYAML(templateJSON)
		if err != nil {
			return "", err
		}

		customTemplateYamls = append([]string{string(templateYaml)}, customTemplateYamls...)
	}
	customTemplateYamls = append([]string{string(baseYaml)}, customTemplateYamls...)

	mergedYaml, err := yaml2.ReplacementMerge(customTemplateYamls...)
	if err != nil {
		return "", err
	}

	return mergedYaml, nil
}