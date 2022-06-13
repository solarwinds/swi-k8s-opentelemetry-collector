package filterdatapointsprocessor

import (
	"context"
	"testing"

	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.uber.org/zap"
	"google.golang.org/protobuf/testing/protocmp"

	internaldata "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/opencensus"
)

func TestFilterDataPointsProcessor(t *testing.T) {
	for _, test := range standardTests {
		t.Run(test.name, func(t *testing.T) {
			next := new(consumertest.MetricsSink)

			p := newFilterDataPointsProcessor(zap.NewExample(), test.transforms)

			mtp, err := processorhelper.NewMetricsProcessor(
				&Config{
					ProcessorSettings: config.NewProcessorSettings(config.NewComponentID(typeStr)),
				},
				next,
				p.processMetrics,
				processorhelper.WithCapabilities(consumerCapabilities))
			require.NoError(t, err)

			caps := mtp.Capabilities()
			assert.Equal(t, true, caps.MutatesData)
			ctx := context.Background()

			// process
			cErr := mtp.ConsumeMetrics(context.Background(), internaldata.OCToMetrics(nil, nil, test.in))
			assert.NoError(t, cErr)

			// get and check results
			got := next.AllMetrics()
			require.Equal(t, 1, len(got))
			_, _, actualOutMetrics := internaldata.ResourceMetricsToOC(got[0].ResourceMetrics().At(0))
			require.Equal(t, len(test.out), len(actualOutMetrics))

			for idx, out := range test.out {
				actualOut := actualOutMetrics[idx]
				if diff := cmp.Diff(actualOut, out, protocmp.Transform()); diff != "" {
					t.Errorf("Unexpected difference:\n%v", diff)
				}
			}

			assert.NoError(t, mtp.Shutdown(ctx))
		})
	}
}

func BenchmarkFilterDataPointsProcessorRenameMetrics(b *testing.B) {
	const metricCount = 1000

	transforms := []internalTransform{
		{
			MetricIncludeFilter: internalFilterStrict{include: "metric"},
			Action:              Insert,
			NewName:             "new/metric1",
		},
	}

	in := make([]*metricspb.Metric, metricCount)
	for i := 0; i < metricCount; i++ {
		in[i] = metricBuilder().setName("metric1").build()
	}
	p := newFilterDataPointsProcessor(nil, transforms)
	mtp, _ := processorhelper.NewMetricsProcessor(&Config{}, consumertest.NewNop(), p.processMetrics)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mtp.ConsumeMetrics(context.Background(), internaldata.OCToMetrics(nil, nil, in))
	}
}
