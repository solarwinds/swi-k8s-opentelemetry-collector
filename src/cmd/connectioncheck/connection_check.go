package connectioncheck

import (
	"context"
	"log"
	"time"

	"github.com/solarwinds/swi-k8s-opentelemetry-collector/metadata"
	"github.com/spf13/cobra"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func sendTestMessage(endpoint, apiToken, clusterUid string, insecure bool) {
	ctx := context.Background()
	otel.SetErrorHandler(new(OtelErrorHandler))

	exporterOptions := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithHeaders(map[string]string{"Authorization": "Bearer " + apiToken}),
		otlploggrpc.WithCompressor("gzip"),
	}

	if insecure {
		exporterOptions = append(exporterOptions, otlploggrpc.WithInsecure())
	}

	exporter, err := otlploggrpc.New(ctx, exporterOptions...)
	if err != nil {
		log.Fatalf("ERROR: Failed to create log exporter\nDETAILS: %s", err)
	}

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)),
		sdklog.WithResource(resource.NewWithAttributes("", attribute.String("sw.k8s.cluster.uid", clusterUid))),
	)
	defer loggerProvider.Shutdown(ctx)

	logger := loggerProvider.Logger(metadata.AppName, otellog.WithInstrumentationVersion(metadata.AppVersion))

	record := otellog.Record{}
	record.SetSeverityText("INFO")
	record.SetBody(otellog.StringValue("otel-endpoint-check successful"))
	record.SetTimestamp(time.Now())

	logger.Emit(ctx, record)
	log.Print("Connection check was successful")
}

type OtelErrorHandler struct{}

func (d *OtelErrorHandler) Handle(err error) {
	switch status.Code(err) {
	case codes.Unauthenticated:
		log.Fatalf("ERROR: A valid token is not set\nDETAILS: %s", err)
	case codes.Unavailable:
		log.Fatalf("ERROR: The target endpoint is not available\nDETAILS: %s", err)
	default:
		log.Fatalf("ERROR: %s", err)
	}
}

func NewCommand() *cobra.Command {
	var clusterUid, endpoint, apiToken string
	var insecure bool

	testCommand := &cobra.Command{
		Use:   "test-connection",
		Short: "Sends a single log to the provided endpoint",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			sendTestMessage(endpoint, apiToken, clusterUid, insecure)
		},
	}
	testCommand.Flags().StringVar(&clusterUid, "clusteruid", "", "")
	testCommand.Flags().StringVar(&endpoint, "endpoint", "", "")
	testCommand.Flags().StringVar(&apiToken, "apitoken", "", "")
	testCommand.Flags().BoolVar(&insecure, "insecure", false, "")
	testCommand.MarkFlagRequired("clusteruid")
	testCommand.MarkFlagRequired("endpoint")
	testCommand.MarkFlagRequired("apitoken")

	return testCommand
}
