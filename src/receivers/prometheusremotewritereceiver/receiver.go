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
	"net"
	"net/http"
	"strconv"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"

	"go.opentelemetry.io/collector/config/confighttp"

	"github.com/prometheus/prometheus/prompb"
	promremote "github.com/prometheus/prometheus/storage/remote"
)

const (
	jsonContentType     = "application/json"
	fallbackContentType = "application/json"
)

var (
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

	if config.HTTP != nil {
		r.httpMux = http.NewServeMux()
	}

	return r, nil

}

// Start is invoked during service startup.
func (r *prometheusRemoteWriteReceiver) Start(_ context.Context, host component.Host) error {
	return r.startProtocolServers(host)
}

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
	r.settings.Logger.Info("Trying to register handler")
	if mc == nil {
		return componenterror.ErrNilNextConsumer
	}

	if r.httpMux != nil {
		r.settings.Logger.Info("Registering handler")
		r.httpMux.HandleFunc("/receive", func(resp http.ResponseWriter, req *http.Request) {
			r.settings.Logger.Info("Received request")
			promreq, err := promremote.DecodeWriteRequest(req.Body)
			ctx := context.WithValue(context.Background(), "request", r)

			if err != nil {
				fmt.Println(err)
				resp.WriteHeader(http.StatusBadRequest)
				return
			}

			md := r.promMetricsToOtelMetrics(promreq)
			err = mc.ConsumeMetrics(ctx, md)
			if err != nil {
				fmt.Println(err)
				resp.WriteHeader(http.StatusBadRequest)
				return
			}

			resp.WriteHeader(http.StatusAccepted)
		})
	}
	return nil
}

func (r *prometheusRemoteWriteReceiver) promMetricsToOtelMetrics(req *prompb.WriteRequest) pmetric.Metrics {
	md := pmetric.NewMetrics()
	/*
		TODO - construct Otel metric here
	*/
	for _, metadata := range req.Metadata {
		r.settings.Logger.Info("Name: " + metadata.MetricFamilyName + ", Type: " + metadata.Type.String() + ", Unit: " + metadata.Unit)
	}
	for _, ts := range req.Timeseries {

		for _, l := range ts.Labels {
			r.settings.Logger.Info("Label: " + l.Name + ", Value: " + l.Value)
		}

		for _, s := range ts.Samples {
			r.settings.Logger.Info("Sample: Timestamp: " + strconv.FormatInt(s.Timestamp, 10) + ", Value: " + fmt.Sprintf("%f", s.Value))
		}
	}
	return md
}
