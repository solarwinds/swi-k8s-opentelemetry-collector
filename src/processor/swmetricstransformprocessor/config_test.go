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

	"github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swmetricstransformprocessor/internal/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadingFullConfig(t *testing.T) {
	tests := []struct {
		configFile string
		id         component.ID
		expCfg     *Config
	}{
		{
			configFile: "config_full.yaml",
			id:         component.NewID(metadata.Type),
			expCfg: &Config{
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
			id:         component.NewIDWithName(metadata.Type, "multiple"),
			expCfg: &Config{
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
								Action:               "filter_datapoints",
								DataPointValue:       1,
								DataPointValueAction: Include,
							},
						},
					},
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "name5",
							MatchType: "strict",
						},
						Action:  "insert",
						NewName: "new_name_copy_4",
						Operations: []Operation{
							{
								Action:               "filter_datapoints",
								DataPointValue:       1,
								DataPointValueAction: Exclude,
							},
						},
					},
				},
			},
		},
		{
			configFile: "config_deprecated.yaml",
			id:         component.NewID(metadata.Type),
			expCfg: &Config{
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
		t.Run(test.id.String(), func(t *testing.T) {

			cm, err := confmaptest.LoadConf(filepath.Join("testdata", test.configFile))
			require.NoError(t, err)

			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(test.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))

			assert.NoError(t, component.ValidateConfig(cfg))
			assert.Equal(t, test.expCfg, cfg)
		})
	}
}
