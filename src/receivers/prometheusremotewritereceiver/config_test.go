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

package prometheusremotewritereceiver

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/service/servicetest"
)

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.NoError(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	cfg, err := servicetest.LoadConfigAndValidate(filepath.Join("testdata", "config.yaml"), factories)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, len(cfg.Receivers), 12)

	assert.Equal(t, cfg.Receivers[config.NewComponentID(typeStr)], factory.CreateDefaultConfig())

	defaultOnlyGRPC := factory.CreateDefaultConfig().(*Config)
	defaultOnlyGRPC.SetIDName("only_grpc")
	defaultOnlyGRPC.HTTP = nil
	assert.Equal(t, cfg.Receivers[config.NewComponentIDWithName(typeStr, "only_grpc")], defaultOnlyGRPC)

	defaultOnlyHTTP := factory.CreateDefaultConfig().(*Config)
	defaultOnlyHTTP.SetIDName("only_http")
	defaultOnlyHTTP.GRPC = nil
	assert.Equal(t, cfg.Receivers[config.NewComponentIDWithName(typeStr, "only_http")], defaultOnlyHTTP)

	defaultOnlyHTTPNull := factory.CreateDefaultConfig().(*Config)
	defaultOnlyHTTPNull.SetIDName("only_http_null")
	defaultOnlyHTTPNull.GRPC = nil
	assert.Equal(t, cfg.Receivers[config.NewComponentIDWithName(typeStr, "only_http_null")], defaultOnlyHTTPNull)

	defaultOnlyHTTPEmptyMap := factory.CreateDefaultConfig().(*Config)
	defaultOnlyHTTPEmptyMap.SetIDName("only_http_empty_map")
	defaultOnlyHTTPEmptyMap.GRPC = nil
	assert.Equal(t, cfg.Receivers[config.NewComponentIDWithName(typeStr, "only_http_empty_map")], defaultOnlyHTTPEmptyMap)

	assert.Equal(t, cfg.Receivers[config.NewComponentIDWithName(typeStr, "customname")],
		&Config{
			ReceiverSettings: config.NewReceiverSettings(config.NewComponentIDWithName(typeStr, "customname")),
			Protocols: Protocols{
				GRPC: &configgrpc.GRPCServerSettings{
					NetAddr: confignet.NetAddr{
						Endpoint:  "localhost:9090",
						Transport: "tcp",
					},
					ReadBufferSize: 512 * 1024,
				},
			},
		})

	assert.Equal(t, cfg.Receivers[config.NewComponentIDWithName(typeStr, "keepalive")],
		&Config{
			ReceiverSettings: config.NewReceiverSettings(config.NewComponentIDWithName(typeStr, "keepalive")),
			Protocols: Protocols{
				GRPC: &configgrpc.GRPCServerSettings{
					NetAddr: confignet.NetAddr{
						Endpoint:  "0.0.0.0:4317",
						Transport: "tcp",
					},
					ReadBufferSize: 512 * 1024,
					Keepalive: &configgrpc.KeepaliveServerConfig{
						ServerParameters: &configgrpc.KeepaliveServerParameters{
							MaxConnectionIdle:     11 * time.Second,
							MaxConnectionAge:      12 * time.Second,
							MaxConnectionAgeGrace: 13 * time.Second,
							Time:                  30 * time.Second,
							Timeout:               5 * time.Second,
						},
						EnforcementPolicy: &configgrpc.KeepaliveEnforcementPolicy{
							MinTime:             10 * time.Second,
							PermitWithoutStream: true,
						},
					},
				},
			},
		})

	assert.Equal(t, cfg.Receivers[config.NewComponentIDWithName(typeStr, "msg-size-conc-connect-max-idle")],
		&Config{
			ReceiverSettings: config.NewReceiverSettings(config.NewComponentIDWithName(typeStr, "msg-size-conc-connect-max-idle")),
			Protocols: Protocols{
				GRPC: &configgrpc.GRPCServerSettings{
					NetAddr: confignet.NetAddr{
						Endpoint:  "0.0.0.0:4317",
						Transport: "tcp",
					},
					MaxRecvMsgSizeMiB:    32,
					MaxConcurrentStreams: 16,
					ReadBufferSize:       1024,
					WriteBufferSize:      1024,
					Keepalive: &configgrpc.KeepaliveServerConfig{
						ServerParameters: &configgrpc.KeepaliveServerParameters{
							MaxConnectionIdle: 10 * time.Second,
						},
					},
				},
			},
		})

	// NOTE: Once the config loader checks for the files existence, this test may fail and require
	// 	use of fake cert/key for test purposes.
	assert.Equal(t, cfg.Receivers[config.NewComponentIDWithName(typeStr, "tlscredentials")],
		&Config{
			ReceiverSettings: config.NewReceiverSettings(config.NewComponentIDWithName(typeStr, "tlscredentials")),
			Protocols: Protocols{
				GRPC: &configgrpc.GRPCServerSettings{
					NetAddr: confignet.NetAddr{
						Endpoint:  "0.0.0.0:4317",
						Transport: "tcp",
					},
					TLSSetting: &configtls.TLSServerSetting{
						TLSSetting: configtls.TLSSetting{
							CertFile: "test.crt",
							KeyFile:  "test.key",
						},
					},
					ReadBufferSize: 512 * 1024,
				},
				HTTP: &confighttp.HTTPServerSettings{
					Endpoint: "0.0.0.0:4318",
					TLSSetting: &configtls.TLSServerSetting{
						TLSSetting: configtls.TLSSetting{
							CertFile: "test.crt",
							KeyFile:  "test.key",
						},
					},
				},
			},
		})

	assert.Equal(t, cfg.Receivers[config.NewComponentIDWithName(typeStr, "cors")],
		&Config{
			ReceiverSettings: config.NewReceiverSettings(config.NewComponentIDWithName(typeStr, "cors")),
			Protocols: Protocols{
				HTTP: &confighttp.HTTPServerSettings{
					Endpoint: "0.0.0.0:4318",
					CORS: &confighttp.CORSSettings{
						AllowedOrigins: []string{"https://*.test.com", "https://test.com"},
						MaxAge:         7200,
					},
				},
			},
		})

	assert.Equal(t, cfg.Receivers[config.NewComponentIDWithName(typeStr, "corsheader")],
		&Config{
			ReceiverSettings: config.NewReceiverSettings(config.NewComponentIDWithName(typeStr, "corsheader")),
			Protocols: Protocols{
				HTTP: &confighttp.HTTPServerSettings{
					Endpoint: "0.0.0.0:4318",
					CORS: &confighttp.CORSSettings{
						AllowedOrigins: []string{"https://*.test.com", "https://test.com"},
						AllowedHeaders: []string{"ExampleHeader"},
					},
				},
			},
		})

	assert.Equal(t, cfg.Receivers[config.NewComponentIDWithName(typeStr, "uds")],
		&Config{
			ReceiverSettings: config.NewReceiverSettings(config.NewComponentIDWithName(typeStr, "uds")),
			Protocols: Protocols{
				GRPC: &configgrpc.GRPCServerSettings{
					NetAddr: confignet.NetAddr{
						Endpoint:  "/tmp/grpc_otlp.sock",
						Transport: "unix",
					},
					ReadBufferSize: 512 * 1024,
				},
				HTTP: &confighttp.HTTPServerSettings{
					Endpoint: "/tmp/http_otlp.sock",
					// Transport: "unix",
				},
			},
		})
}

func TestFailedLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.NoError(t, err)

	factory := NewFactory()
	factories.Receivers[typeStr] = factory
	_, err = servicetest.LoadConfigAndValidate(filepath.Join("testdata", "typo_default_proto_config.yaml"), factories)
	assert.EqualError(t, err, "error reading receivers configuration for \"otlp\": 1 error(s) decoding:\n\n* 'protocols' has invalid keys: htttp")

	_, err = servicetest.LoadConfigAndValidate(filepath.Join("testdata", "bad_proto_config.yaml"), factories)
	assert.EqualError(t, err, "error reading receivers configuration for \"otlp\": 1 error(s) decoding:\n\n* 'protocols' has invalid keys: thrift")

	_, err = servicetest.LoadConfigAndValidate(filepath.Join("testdata", "bad_no_proto_config.yaml"), factories)
	assert.EqualError(t, err, "receiver \"otlp\" has invalid configuration: must specify at least one protocol when using the OTLP receiver")

	_, err = servicetest.LoadConfigAndValidate(filepath.Join("testdata", "bad_empty_config.yaml"), factories)
	assert.EqualError(t, err, "error reading receivers configuration for \"otlp\": empty config for OTLP receiver")
}
