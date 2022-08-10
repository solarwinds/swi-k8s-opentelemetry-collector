// Copyright 2022 SolarWinds Worldwide, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Source: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/metricstransformprocessor
// Changes customizing the original processor:
//	- removal of actions: toggle_scalar_data_type, experimental_scale_value, aggregate_labels, aggregate_label_values
//	- add custom action "filter_datapoints"
//	- rename types and functions to match the processor name

package swmetricstransformprocessor

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/service/servicetest"
)

func TestLoadingFullConfig(t *testing.T) {
	tests := []struct {
		configFile string
		filterName config.ComponentID
		expCfg     *Config
	}{
		{
			configFile: "config_full.yaml",
			filterName: config.NewComponentID(typeStr),
			expCfg: &Config{
				ProcessorSettings: config.NewProcessorSettings(config.NewComponentID(typeStr)),
				Transforms: []Transform{
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "name",
							MatchType: "",
						},
						Action:  "update",
						NewName: "new_name",
					},
				},
			},
		},
		{
			configFile: "config_full.yaml",
			filterName: config.NewComponentIDWithName(typeStr, "multiple"),
			expCfg: &Config{
				ProcessorSettings: config.NewProcessorSettings(config.NewComponentIDWithName(typeStr, "multiple")),
				Transforms: []Transform{
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "name1",
							MatchType: "strict",
						},
						Action:  "insert",
						NewName: "new_name",
						Operations: []Operation{
							{
								Action:   "add_label",
								NewLabel: "my_label",
								NewValue: "my_value",
							},
						},
					},
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "new_name",
							MatchType: "strict",
							MatchLabels: map[string]string{
								"my_label": "my_value",
							},
						},
						Action:  "insert",
						NewName: "new_name_copy_1",
					},
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "new_name",
							MatchType: "regexp",
							MatchLabels: map[string]string{
								"my_label": ".*label",
							},
						},
						Action:  "insert",
						NewName: "new_name_copy_2",
					},
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "name3",
							MatchType: "strict",
						},
						Action: "update",
						Operations: []Operation{
							{
								Action:     "delete_label_value",
								Label:      "my_label",
								LabelValue: "delete_me",
							},
						},
					},
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "name4",
							MatchType: "strict",
						},
						Action:  "insert",
						NewName: "new_name_copy_3",
						Operations: []Operation{
							{
								Action:         "filter_datapoints",
								DataPointValue: 1,
							},
						},
					},
				},
			},
		},
		{
			configFile: "config_deprecated.yaml",
			filterName: config.NewComponentID(typeStr),
			expCfg: &Config{
				ProcessorSettings: config.NewProcessorSettings(config.NewComponentID(typeStr)),
				Transforms: []Transform{
					{
						MetricName: "old_name",
						Action:     Update,
						NewName:    "new_name",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.filterName.String(), func(t *testing.T) {

			factories, err := componenttest.NopFactories()
			assert.NoError(t, err)

			factory := NewFactory()
			factories.Processors[typeStr] = factory
			cfg, err := servicetest.LoadConfigAndValidate(filepath.Join("testdata", test.configFile), factories)
			assert.NoError(t, err)
			require.NotNil(t, cfg)
			assert.Equal(t, test.expCfg, cfg.Processors[test.filterName])
		})
	}
}
