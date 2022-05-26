package prometheustypeconverterprocessor

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
							Include: "name",
						},
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
