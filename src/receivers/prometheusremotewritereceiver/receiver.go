// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prometheusremotewritereceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricsgenerationprocessor"

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"

	"go.opentelemetry.io/collector/config/confighttp"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	promremote "github.com/prometheus/prometheus/storage/remote"
)

const (
	jsonContentType     = "application/json"
	fallbackContentType = "application/json"
)

var (
	//jsEncoder     = &jsonEncoder{}
	//jsonMarshaler = &jsonpb.Marshaler{}
	fallbackMsg = []byte(`{"code": 13, "message": "failed to marshal error message"}`)
)

type prometheusRemoteWriteReceiver struct {
	cfg        *Config
	httpMux    *http.ServeMux
	serverHTTP *http.Server

	shutdownWG sync.WaitGroup

	settings component.ReceiverCreateSettings
}

func newMetricsReceiver(ctx context.Context,
	config *Config,
	params component.ReceiverCreateSettings,
	consumer consumer.Metrics,
) (component.MetricsReceiver, error) {
	r := &prometheusRemoteWriteReceiver{
		cfg:      config,
		settings: params,
	}

	/*
		if cfg.HTTP != nil {
			r.httpMux = http.NewServeMux()
		}*/

	return r, nil

}

// Start is invoked during service startup.
func (r *prometheusRemoteWriteReceiver) Start(_ context.Context, host component.Host) error {
	return r.startProtocolServers(host)
}

/*
// processMetrics implements the ProcessMetricsFunc type.
func (mgp *prometheusRemoteWriteReceiver) processMetrics(_ context.Context, md pdata.Metrics) (pdata.Metrics, error) {
	resourceMetricsSlice := md.ResourceMetrics()

	for i := 0; i < resourceMetricsSlice.Len(); i++ {
		rm := resourceMetricsSlice.At(i)
		nameToMetricMap := getNameToMetricMap(rm)
		fmt.Println("Processing metric")

		for _, rule := range mgp.rules {
			operand2 := float64(0)
			_, ok := nameToMetricMap[rule.metric1]
			if !ok {
				fmt.Println("Missing first metric", zap.String("metric_name", rule.metric1))
				continue
			}

			if rule.ruleType == string(calculate) {
				metric2, ok := nameToMetricMap[rule.metric2]
				if !ok {
					fmt.Println("Missing second metric", zap.String("metric_name", rule.metric2))
					continue
				}
				operand2 = getMetricValue(metric2)
				if operand2 <= 0 {
					continue
				}

			} else if rule.ruleType == string(scale) {
				operand2 = rule.scaleBy
			}
			generateMetrics(rm, operand2, rule, mgp.logger)
		}
	}
	return md, nil
}
*/

func (r *prometheusRemoteWriteReceiver) startHTTPServer(cfg *confighttp.HTTPServerSettings, host component.Host) error {
	r.settings.Logger.Info("Starting HTTP server on endpoint " + cfg.Endpoint)
	var hln net.Listener
	hln, err := cfg.ToListener()
	if err != nil {
		return err
	}
	r.shutdownWG.Add(1)
	go func() {
		defer r.shutdownWG.Done()

		if errHTTP := r.serverHTTP.Serve(hln); errHTTP != http.ErrServerClosed {
			host.ReportFatalError(errHTTP)
		}
	}()
	return nil
}

func (r *prometheusRemoteWriteReceiver) startProtocolServers(host component.Host) error {
	var err error
	if r.cfg.HTTP != nil {
		r.serverHTTP, err = r.cfg.HTTP.ToServer(
			host,
			r.settings.TelemetrySettings,
			r.httpMux,
			//confighttp.WithErrorHandler(errorHandler),
		)
		if err != nil {
			return err
		}

		err = r.startHTTPServer(r.cfg.HTTP, host)
		if err != nil {
			return err
		}
	}

	return err
}

// Shutdown is invoked during service shutdown.
func (r *prometheusRemoteWriteReceiver) Shutdown(ctx context.Context) error {
	var err error

	if r.serverHTTP != nil {
		err = r.serverHTTP.Shutdown(ctx)
	}

	r.shutdownWG.Wait()
	return err
}

func (r *prometheusRemoteWriteReceiver) registerMetricsConsumer(mc consumer.Metrics) error {
	if mc == nil {
		return componenterror.ErrNilNextConsumer
	}

	//r.metricsReceiver = metrics.New(r.cfg.ID(), mc, r.settings)
	if r.httpMux != nil {
		r.httpMux.HandleFunc("/receive", func(resp http.ResponseWriter, req *http.Request) {
			//ctx := context.WithValue(context.Background(), "request", r)

			compressed, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Println(err)
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}

			data, err := processRequestData(compressed)
			if err != nil {
				fmt.Println(err)
				resp.WriteHeader(http.StatusBadRequest)
				return
			}

			//convertedData := adapter.PromDataToAppOpticsMeasurements(&data)
			//msg := fmt.Sprintf("measurements received - %d", len(convertedData))
			fmt.Println(data)

			/*md := req.Metrics()
			dataPointCount := md.DataPointCount()
			if dataPointCount == 0 {
				return pmetricotlp.NewResponse(), nil
			}*/

			//err := mc.ConsumeMetrics(ctx, md)

			/*emptyMap := make(map[string]string)

			batch := appoptics.NewMeasurementsBatch(convertedData, &emptyMap)
			if len(batch.Measurements) == 0 {
				fmt.Println("Skipping payload with zero measurements")
			} else {
				resp, err := aoClient.MeasurementsService().Create(batch)
				if err != nil {
					fmt.Println("Error submitting metrics to Appoptics", err)
				}
				if resp != nil && resp.StatusCode != http.StatusAccepted {
					fmt.Println(resp.StatusCode)
				}
			}*/
			// We do not want to propagate errors upstream yet
			resp.WriteHeader(http.StatusAccepted)
		})
	}
	return nil
}

func processRequestData(reqBytes []byte) (promremote.WriteRequest, error) {
	var req promremote.WriteRequest
	reqBuf, err := snappy.Decode(nil, reqBytes)
	if err != nil {
		return req, err
	}

	if err := proto.Unmarshal(reqBuf, &req); err != nil {
		return req, err
	}
	return req, nil
}

/*
func promDataToOtelMetrics(req *promremote.WriteRequest) (promremote.WriteRequest, error) {
	var req promremote.WriteRequest
	reqBuf, err := snappy.Decode(nil, reqBytes)
	if err != nil {
		return req, err
	}

	if err := proto.Unmarshal(reqBuf, &req); err != nil {
		return req, err
	}
	return req, nil
}*/

/*
// errorHandler encodes the HTTP error message inside a rpc.Status message as required
// by the OTLP protocol.
func errorHandler(w http.ResponseWriter, r *http.Request, errMsg string, statusCode int) {
	s := errorMsgToStatus(errMsg, statusCode)
	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case jsonContentType:
		writeStatusResponse(w, jsEncoder, statusCode, s.Proto())
		return
	}
	writeResponse(w, fallbackContentType, http.StatusInternalServerError, fallbackMsg)
}*/
/*
func errorMsgToStatus(errMsg string, statusCode int) *status.Status {
	if statusCode == http.StatusBadRequest {
		return status.New(codes.InvalidArgument, errMsg)
	}
	return status.New(codes.Unknown, errMsg)
}

func writeResponse(w http.ResponseWriter, contentType string, statusCode int, msg []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	// Nothing we can do with the error if we cannot write to the response.
	_, _ = w.Write(msg)
}

func writeStatusResponse(w http.ResponseWriter, encoder encoder, statusCode int, rsp *spb.Status) {
	msg, err := encoder.marshalStatus(rsp)
	if err != nil {
		writeResponse(w, fallbackContentType, http.StatusInternalServerError, fallbackMsg)
		return
	}

	writeResponse(w, encoder.contentType(), statusCode, msg)
}*/
