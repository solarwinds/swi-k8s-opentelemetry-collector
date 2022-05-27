package prometheustypeconverterprocessor

import (
	"go.opentelemetry.io/collector/config"
)

const (
	// IncludeFieldName is the mapstructure field name for Include field
	IncludeFieldName = "include"

	// ConvertTypeFieldName is the mapstructure field name for ConvertType field
	ConvertTypeFieldName = "convert_type"
)

// Config defines configuration for Resource processor.
type Config struct {
	config.ProcessorSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct

	// Transform specifies a list of transforms on metrics with each transform focusing on one metric.
	Transforms []Transform `mapstructure:"transforms"`
}

// Transform defines the transformation applied to the specific metric
type Transform struct {

	// --- SPECIFY WHICH METRIC(S) TO MATCH ---

	// MetricIncludeFilter is used to select the metric(s) to operate on.
	// REQUIRED
	MetricIncludeFilter FilterConfig `mapstructure:",squash"`

	// ConvertType determines to what type it should be converted
	ConvertType ConvertType `mapstructure:"convert_type"`
}

type FilterConfig struct {
	// Include specifies the metric(s) to operate on.
	Include string `mapstructure:"include"`
}

// ConvertType is the enum which indicates to what type the metric is converted to.
type ConvertType string

const (
	// SumConvertType is the ConvertType indicating that type should be converted to Sum.
	SumConvertType ConvertType = "sum"
)

var convertTypes = []ConvertType{SumConvertType}

func (ct ConvertType) isValid() bool {
	for _, convertType := range convertTypes {
		if ct == convertType {
			return true
		}
	}

	return false
}
