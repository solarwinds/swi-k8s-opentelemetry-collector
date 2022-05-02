// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//lint:file-ignore U1000 t.Skip() flaky test causes unused function warning.

package prometheusremotewritereceiver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	commonpb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/common/v1"
	agentmetricspb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/metrics/v1"
	agenttracepb "github.com/census-instrumentation/opencensus-proto/gen-go/agent/trace/v1"
	metricspb "github.com/census-instrumentation/opencensus-proto/gen-go/metrics/v1"
	resourcepb "github.com/census-instrumentation/opencensus-proto/gen-go/resource/v1"
	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/obsreport/obsreporttest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/common/testutil"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/opencensus"
)

var ocReceiverID = config.NewComponentIDWithName(typeStr, "receiver_test")

func TestGrpcGateway_endToEnd(t *testing.T) {
	addr := testutil.GetAvailableLocalAddress(t)

	// Set the buffer count to 1 to make it flush the test span immediately.
	sink := new(consumertest.TracesSink)
	ocr, err := newOpenCensusReceiver(ocReceiverID, "tcp", addr, sink, nil, componenttest.NewNopReceiverCreateSettings())
	require.NoError(t, err, "Failed to create trace receiver: %v", err)

	err = ocr.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err, "Failed to start trace receiver: %v", err)
	t.Cleanup(func() { require.NoError(t, ocr.Shutdown(context.Background())) })

	// TODO(songy23): make starting server deterministic
	// Wait for the servers to start
	<-time.After(10 * time.Millisecond)

	url := fmt.Sprintf("http://%s/v1/trace", addr)

	// Verify that CORS is not enabled by default, but that it gives a method not allowed error.
	verifyCorsResp(t, url, "origin.com", http.StatusMethodNotAllowed, false)

	traceJSON := []byte(`
    {
       "node":{"identifier":{"hostName":"testHost"}},
       "spans":[
          {
              "traceId":"W47/95gDgQPSabYzgT/GDA==",
              "spanId":"7uGbfsPBsXM=",
              "name":{"value":"testSpan"},
              "startTime":"2018-12-13T14:51:00Z",
              "endTime":"2018-12-13T14:51:01Z",
              "attributes": {
                "attributeMap": {
                  "attr1": {"intValue": "55"}
                }
              }
          }
       ]
    }`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(traceJSON))
	require.NoError(t, err, "Error creating trace POST request: %v", err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Error posting trace to grpc-gateway server: %v", err)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Error reading response from trace grpc-gateway, %v", err)
	}
	respStr := string(respBytes)

	err = resp.Body.Close()
	if err != nil {
		t.Errorf("Error closing response body, %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Unexpected status from trace grpc-gateway: %v", resp.StatusCode)
	}

	if respStr != "" {
		t.Errorf("Got unexpected response from trace grpc-gateway: %v", respStr)
	}

	got := sink.AllTraces()
	require.Len(t, got, 1)
	require.Equal(t, 1, got[0].ResourceSpans().Len())
	gotNode, gotResource, gotSpans := opencensus.ResourceSpansToOC(got[0].ResourceSpans().At(0))

	wantNode := &commonpb.Node{Identifier: &commonpb.ProcessIdentifier{HostName: "testHost"}}
	wantResource := &resourcepb.Resource{}
	wantSpans := []*tracepb.Span{
		{
			TraceId:   []byte{0x5B, 0x8E, 0xFF, 0xF7, 0x98, 0x3, 0x81, 0x3, 0xD2, 0x69, 0xB6, 0x33, 0x81, 0x3F, 0xC6, 0xC},
			SpanId:    []byte{0xEE, 0xE1, 0x9B, 0x7E, 0xC3, 0xC1, 0xB1, 0x73},
			Name:      &tracepb.TruncatableString{Value: "testSpan"},
			StartTime: timestamppb.New(time.Unix(1544712660, 0).UTC()),
			EndTime:   timestamppb.New(time.Unix(1544712661, 0).UTC()),
			Attributes: &tracepb.Span_Attributes{
				AttributeMap: map[string]*tracepb.AttributeValue{
					"attr1": {
						Value: &tracepb.AttributeValue_IntValue{IntValue: 55},
					},
				},
			},
			Status: &tracepb.Status{},
		},
	}
	assert.True(t, proto.Equal(wantNode, gotNode))
	assert.True(t, proto.Equal(wantResource, gotResource))
	require.Len(t, wantSpans, 1)
	require.Len(t, gotSpans, 1)
	assert.EqualValues(t, wantSpans[0], gotSpans[0])
}

func TestTraceGrpcGatewayCors_endToEnd(t *testing.T) {
	addr := testutil.GetAvailableLocalAddress(t)
	corsOrigins := []string{"allowed-*.com"}

	ocr, err := newOpenCensusReceiver(ocReceiverID, "tcp", addr, consumertest.NewNop(), nil, componenttest.NewNopReceiverCreateSettings(), withCorsOrigins(corsOrigins))
	require.NoError(t, err, "Failed to create trace receiver: %v", err)
	t.Cleanup(func() { require.NoError(t, ocr.Shutdown(context.Background())) })

	err = ocr.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err, "Failed to start trace receiver: %v", err)

	// TODO(songy23): make starting server deterministic
	// Wait for the servers to start
	<-time.After(10 * time.Millisecond)

	url := fmt.Sprintf("http://%s/v1/trace", addr)

	// Verify allowed domain gets responses that allow CORS.
	verifyCorsResp(t, url, "allowed-origin.com", http.StatusNoContent, true)

	// Verify disallowed domain gets responses that disallow CORS.
	verifyCorsResp(t, url, "disallowed-origin.com", http.StatusNoContent, false)
}

func TestMetricsGrpcGatewayCors_endToEnd(t *testing.T) {
	addr := testutil.GetAvailableLocalAddress(t)
	corsOrigins := []string{"allowed-*.com"}

	ocr, err := newOpenCensusReceiver(ocReceiverID, "tcp", addr, nil, consumertest.NewNop(), componenttest.NewNopReceiverCreateSettings(), withCorsOrigins(corsOrigins))
	require.NoError(t, err, "Failed to create metrics receiver: %v", err)
	t.Cleanup(func() { require.NoError(t, ocr.Shutdown(context.Background())) })

	err = ocr.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err, "Failed to start metrics receiver: %v", err)

	// TODO(songy23): make starting server deterministic
	// Wait for the servers to start
	<-time.After(10 * time.Millisecond)

	url := fmt.Sprintf("http://%s/v1/metrics", addr)

	// Verify allowed domain gets responses that allow CORS.
	verifyCorsResp(t, url, "allowed-origin.com", http.StatusNoContent, true)

	// Verify disallowed domain gets responses that disallow CORS.
	verifyCorsResp(t, url, "disallowed-origin.com", http.StatusNoContent, false)
}

func verifyCorsResp(t *testing.T, url string, origin string, wantStatus int, wantAllowed bool) {
	req, err := http.NewRequest("OPTIONS", url, nil)
	require.NoError(t, err, "Error creating trace OPTIONS request: %v", err)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", "POST")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Error sending OPTIONS to grpc-gateway server: %v", err)

	err = resp.Body.Close()
	if err != nil {
		t.Errorf("Error closing OPTIONS response body, %v", err)
	}

	if resp.StatusCode != wantStatus {
		t.Errorf("Unexpected status from OPTIONS: %v", resp.StatusCode)
	}

	gotAllowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	gotAllowMethods := resp.Header.Get("Access-Control-Allow-Methods")

	wantAllowOrigin := ""
	wantAllowMethods := ""
	if wantAllowed {
		wantAllowOrigin = origin
		wantAllowMethods = "POST"
	}

	if gotAllowOrigin != wantAllowOrigin {
		t.Errorf("Unexpected Access-Control-Allow-Origin: %v", gotAllowOrigin)
	}
	if gotAllowMethods != wantAllowMethods {
		t.Errorf("Unexpected Access-Control-Allow-Methods: %v", gotAllowMethods)
	}
}

func TestStopWithoutStartNeverCrashes(t *testing.T) {
	addr := testutil.GetAvailableLocalAddress(t)
	ocr, err := newOpenCensusReceiver(ocReceiverID, "tcp", addr, nil, nil, componenttest.NewNopReceiverCreateSettings())
	require.NoError(t, err, "Failed to create an OpenCensus receiver: %v", err)
	// Stop it before ever invoking Start*.
	require.NoError(t, ocr.Shutdown(context.Background()))
}

func TestNewPortAlreadyUsed(t *testing.T) {
	addr := testutil.GetAvailableLocalAddress(t)
	ln, err := net.Listen("tcp", addr)
	require.NoError(t, err, "failed to listen on %q: %v", addr, err)
	defer ln.Close()

	r, err := newOpenCensusReceiver(ocReceiverID, "tcp", addr, nil, nil, componenttest.NewNopReceiverCreateSettings())
	require.Error(t, err)
	require.Nil(t, r)
}

func TestMultipleStopReceptionShouldNotError(t *testing.T) {
	addr := testutil.GetAvailableLocalAddress(t)
	r, err := newOpenCensusReceiver(ocReceiverID, "tcp", addr, consumertest.NewNop(), consumertest.NewNop(), componenttest.NewNopReceiverCreateSettings())
	require.NoError(t, err)
	require.NotNil(t, r)

	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, r.Shutdown(context.Background()))
}

func TestStartWithoutConsumersShouldFail(t *testing.T) {
	addr := testutil.GetAvailableLocalAddress(t)
	r, err := newOpenCensusReceiver(ocReceiverID, "tcp", addr, nil, nil, componenttest.NewNopReceiverCreateSettings())
	require.NoError(t, err)
	require.NotNil(t, r)

	require.Error(t, r.Start(context.Background(), componenttest.NewNopHost()))
}

func tempSocketName(t *testing.T) string {
	tmpfile, err := ioutil.TempFile("", "sock")
	require.NoError(t, err)
	require.NoError(t, tmpfile.Close())
	socket := tmpfile.Name()
	require.NoError(t, os.Remove(socket))
	return socket
}

func TestReceiveOnUnixDomainSocket_endToEnd(t *testing.T) {
	socketName := tempSocketName(t)
	cbts := consumertest.NewNop()
	r, err := newOpenCensusReceiver(ocReceiverID, "unix", socketName, cbts, nil, componenttest.NewNopReceiverCreateSettings())
	require.NoError(t, err)
	require.NotNil(t, r)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { require.NoError(t, r.Shutdown(context.Background())) })

	// Wait for the servers to start
	<-time.After(10 * time.Millisecond)

	span := `
{
 "node": {
 },
 "spans": [
   {
     "trace_id": "YpsR8/le4OgjwSSxhjlrEg==",
     "span_id": "2CogcbJh7Ko=",
     "socket": {
       "value": "/abc",
       "truncated_byte_count": 0
     },
     "kind": "SPAN_KIND_UNSPECIFIED",
     "start_time": "2020-01-09T11:13:53.187Z",
     "end_time": "2020-01-09T11:13:53.187Z"
	}
 ]
}
`
	c := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
				return net.Dial("unix", socketName)
			},
		},
	}

	response, err := c.Post("http://unix/v1/trace", "application/json", strings.NewReader(span))
	require.NoError(t, err)
	defer response.Body.Close()

	require.Equal(t, 200, response.StatusCode)
}

// TestOCReceiverTrace_HandleNextConsumerResponse checks if the trace receiver
// is returning the proper response (return and metrics) when the next consumer
// in the pipeline reports error. The test changes the responses returned by the
// next trace consumer, checks if data was passed down the pipeline and if
// proper metrics were recorded. It also uses all endpoints supported by the
// trace receiver.
func TestOCReceiverTrace_HandleNextConsumerResponse(t *testing.T) {
	type ingestionStateTest struct {
		okToIngest   bool
		expectedCode codes.Code
	}
	tests := []struct {
		name                         string
		expectedReceivedBatches      int
		expectedIngestionBlockedRPCs int
		ingestionStates              []ingestionStateTest
	}{
		{
			name:                         "IngestTest",
			expectedReceivedBatches:      2,
			expectedIngestionBlockedRPCs: 1,
			ingestionStates: []ingestionStateTest{
				{
					okToIngest:   true,
					expectedCode: codes.OK,
				},
				{
					okToIngest:   false,
					expectedCode: codes.Unknown,
				},
				{
					okToIngest:   true,
					expectedCode: codes.OK,
				},
			},
		},
	}

	addr := testutil.GetAvailableLocalAddress(t)
	msg := &agenttracepb.ExportTraceServiceRequest{
		Node: &commonpb.Node{
			ServiceInfo: &commonpb.ServiceInfo{Name: "test-svc"},
		},
		Spans: []*tracepb.Span{
			{
				TraceId: []byte{
					0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
					0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10,
				},
			},
		},
	}

	exportBidiFn := func(
		t *testing.T,
		cc *grpc.ClientConn,
		msg *agenttracepb.ExportTraceServiceRequest) error {

		acc := agenttracepb.NewTraceServiceClient(cc)
		stream, err := acc.Export(context.Background())
		require.NoError(t, err)
		require.NotNil(t, stream)

		err = stream.Send(msg)
		require.NoError(t, stream.CloseSend())
		if err == nil {
			for {
				if _, err = stream.Recv(); err != nil {
					if err == io.EOF {
						err = nil
					}
					break
				}
			}
		}

		return err
	}

	exporters := []struct {
		receiverID config.ComponentID
		exportFn   func(
			t *testing.T,
			cc *grpc.ClientConn,
			msg *agenttracepb.ExportTraceServiceRequest) error
	}{
		{
			receiverID: config.NewComponentIDWithName(typeStr, "traces"),
			exportFn:   exportBidiFn,
		},
	}
	for _, exporter := range exporters {
		for _, tt := range tests {
			t.Run(tt.name+"/"+exporter.receiverID.String(), func(t *testing.T) {
				testTel, err := obsreporttest.SetupTelemetry()
				require.NoError(t, err)
				defer testTel.Shutdown(context.Background())

				sink := &errOrSinkConsumer{TracesSink: new(consumertest.TracesSink)}

				var opts []ocOption
				ocr, err := newOpenCensusReceiver(exporter.receiverID, "tcp", addr, nil, nil, componenttest.NewNopReceiverCreateSettings(), opts...)
				require.Nil(t, err)
				require.NotNil(t, ocr)

				ocr.traceConsumer = sink
				require.NoError(t, ocr.Start(context.Background(), componenttest.NewNopHost()))
				t.Cleanup(func() { require.NoError(t, ocr.Shutdown(context.Background())) })

				cc, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
				if err != nil {
					t.Errorf("grpc.Dial: %v", err)
				}
				defer cc.Close()

				for _, ingestionState := range tt.ingestionStates {
					if ingestionState.okToIngest {
						sink.SetConsumeError(nil)
					} else {
						sink.SetConsumeError(fmt.Errorf("%q: consumer error", tt.name))
					}

					err = exporter.exportFn(t, cc, msg)

					status, ok := status.FromError(err)
					require.True(t, ok)
					assert.Equal(t, ingestionState.expectedCode, status.Code())
				}

				require.Equal(t, tt.expectedReceivedBatches, len(sink.AllTraces()))
				require.NoError(t, obsreporttest.CheckReceiverTraces(testTel, exporter.receiverID, "grpc", int64(tt.expectedReceivedBatches), int64(tt.expectedIngestionBlockedRPCs)))
			})
		}
	}
}

// TestOCReceiverMetrics_HandleNextConsumerResponse checks if the metrics receiver
// is returning the proper response (return and metrics) when the next consumer
// in the pipeline reports error. The test changes the responses returned by the
// next trace consumer, checks if data was passed down the pipeline and if
// proper metrics were recorded. It also uses all endpoints supported by the
// metrics receiver.
func TestOCReceiverMetrics_HandleNextConsumerResponse(t *testing.T) {
	type ingestionStateTest struct {
		okToIngest   bool
		expectedCode codes.Code
	}
	tests := []struct {
		name                         string
		expectedReceivedBatches      int
		expectedIngestionBlockedRPCs int
		ingestionStates              []ingestionStateTest
	}{
		{
			name:                         "IngestTest",
			expectedReceivedBatches:      2,
			expectedIngestionBlockedRPCs: 1,
			ingestionStates: []ingestionStateTest{
				{
					okToIngest:   true,
					expectedCode: codes.OK,
				},
				{
					okToIngest:   false,
					expectedCode: codes.Unknown,
				},
				{
					okToIngest:   true,
					expectedCode: codes.OK,
				},
			},
		},
	}

	descriptor := &metricspb.MetricDescriptor{
		Name:        "testMetric",
		Description: "metric descriptor",
		Unit:        "1",
		Type:        metricspb.MetricDescriptor_GAUGE_INT64,
	}
	point := &metricspb.Point{
		Timestamp: timestamppb.New(time.Now().UTC()),
		Value: &metricspb.Point_Int64Value{
			Int64Value: int64(1),
		},
	}
	ts := &metricspb.TimeSeries{
		Points: []*metricspb.Point{point},
	}
	metric := &metricspb.Metric{
		MetricDescriptor: descriptor,
		Timeseries:       []*metricspb.TimeSeries{ts},
	}

	addr := testutil.GetAvailableLocalAddress(t)
	msg := &agentmetricspb.ExportMetricsServiceRequest{
		Node: &commonpb.Node{
			ServiceInfo: &commonpb.ServiceInfo{Name: "test-svc"},
		},
		Metrics: []*metricspb.Metric{metric},
	}

	exportBidiFn := func(
		t *testing.T,
		cc *grpc.ClientConn,
		msg *agentmetricspb.ExportMetricsServiceRequest) error {

		acc := agentmetricspb.NewMetricsServiceClient(cc)
		stream, err := acc.Export(context.Background())
		require.NoError(t, err)
		require.NotNil(t, stream)

		err = stream.Send(msg)
		require.NoError(t, stream.CloseSend())
		if err == nil {
			for {
				if _, err = stream.Recv(); err != nil {
					if err == io.EOF {
						err = nil
					}
					break
				}
			}
		}

		return err
	}

	exporters := []struct {
		receiverID config.ComponentID
		exportFn   func(
			t *testing.T,
			cc *grpc.ClientConn,
			msg *agentmetricspb.ExportMetricsServiceRequest) error
	}{
		{
			receiverID: config.NewComponentIDWithName(typeStr, "metrics"),
			exportFn:   exportBidiFn,
		},
	}
	for _, exporter := range exporters {
		for _, tt := range tests {
			t.Run(tt.name+"/"+exporter.receiverID.String(), func(t *testing.T) {
				testTel, err := obsreporttest.SetupTelemetry()
				require.NoError(t, err)
				defer testTel.Shutdown(context.Background())

				sink := &errOrSinkConsumer{MetricsSink: new(consumertest.MetricsSink)}

				var opts []ocOption
				ocr, err := newOpenCensusReceiver(exporter.receiverID, "tcp", addr, nil, nil, componenttest.NewNopReceiverCreateSettings(), opts...)
				require.Nil(t, err)
				require.NotNil(t, ocr)

				ocr.metricsConsumer = sink
				require.Nil(t, ocr.Start(context.Background(), componenttest.NewNopHost()))
				t.Cleanup(func() { require.NoError(t, ocr.Shutdown(context.Background())) })

				cc, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
				if err != nil {
					t.Errorf("grpc.Dial: %v", err)
				}
				defer cc.Close()

				for _, ingestionState := range tt.ingestionStates {
					if ingestionState.okToIngest {
						sink.SetConsumeError(nil)
					} else {
						sink.SetConsumeError(fmt.Errorf("%q: consumer error", tt.name))
					}

					err = exporter.exportFn(t, cc, msg)

					status, ok := status.FromError(err)
					require.True(t, ok)
					assert.Equal(t, ingestionState.expectedCode, status.Code())
				}

				require.Equal(t, tt.expectedReceivedBatches, len(sink.AllMetrics()))
				require.NoError(t, obsreporttest.CheckReceiverMetrics(testTel, exporter.receiverID, "grpc", int64(tt.expectedReceivedBatches), int64(tt.expectedIngestionBlockedRPCs)))
			})
		}
	}
}

func TestInvalidTLSCredentials(t *testing.T) {
	cfg := Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
		GRPCServerSettings: configgrpc.GRPCServerSettings{
			TLSSetting: &configtls.TLSServerSetting{
				TLSSetting: configtls.TLSSetting{
					CertFile: "willfail",
				},
			},
		},
	}
	opt := cfg.buildOptions()
	assert.NotNil(t, opt)

	addr := testutil.GetAvailableLocalAddress(t)
	ocr, err := newOpenCensusReceiver(config.NewComponentIDWithName(typeStr, "invalidtls"), "tcp", addr, nil, nil, componenttest.NewNopReceiverCreateSettings(), opt...)
	assert.NoError(t, err)
	assert.NotNil(t, ocr)

	srv, err := ocr.grpcServer(componenttest.NewNopHost())
	assert.EqualError(t, err, `failed to load TLS config: for auth via TLS, either both certificate and key must be supplied, or neither`)
	assert.Nil(t, srv)
}

type errOrSinkConsumer struct {
	*consumertest.TracesSink
	*consumertest.MetricsSink
	mu           sync.Mutex
	consumeError error // to be returned by ConsumeTraces, if set
}

// SetConsumeError sets an error that will be returned by the Consume function.
func (esc *errOrSinkConsumer) SetConsumeError(err error) {
	esc.mu.Lock()
	defer esc.mu.Unlock()
	esc.consumeError = err
}

func (esc *errOrSinkConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// ConsumeTraces stores traces to this sink.
func (esc *errOrSinkConsumer) ConsumeTraces(ctx context.Context, td pdata.Traces) error {
	esc.mu.Lock()
	defer esc.mu.Unlock()

	if esc.consumeError != nil {
		return esc.consumeError
	}

	return esc.TracesSink.ConsumeTraces(ctx, td)
}

// ConsumeMetrics stores metrics to this sink.
func (esc *errOrSinkConsumer) ConsumeMetrics(ctx context.Context, md pdata.Metrics) error {
	esc.mu.Lock()
	defer esc.mu.Unlock()

	if esc.consumeError != nil {
		return esc.consumeError
	}

	return esc.MetricsSink.ConsumeMetrics(ctx, md)
}

// Reset deletes any stored in the sinks, resets error to nil.
func (esc *errOrSinkConsumer) Reset() {
	esc.mu.Lock()
	defer esc.mu.Unlock()

	esc.consumeError = nil
	if esc.TracesSink != nil {
		esc.TracesSink.Reset()
	}
	if esc.MetricsSink != nil {
		esc.MetricsSink.Reset()
	}
}
