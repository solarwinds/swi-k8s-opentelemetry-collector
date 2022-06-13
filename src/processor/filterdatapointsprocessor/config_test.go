package filterdatapointsprocessor

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
