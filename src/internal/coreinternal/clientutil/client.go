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

// Source: https://github.com/open-telemetry/opentelemetry-collector-contrib

package clientutil // import "github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/clientutil"

import (
	"net"
	"strings"

	"go.opentelemetry.io/collector/client"
)

// Address returns the address of the client connecting to the collector.
func Address(client client.Info) string {
	if client.Addr == nil {
		return ""
	}
	switch addr := client.Addr.(type) {
	case *net.UDPAddr:
		return addr.IP.String()
	case *net.TCPAddr:
		return addr.IP.String()
	case *net.IPAddr:
		return addr.IP.String()
	}

	// If this is not a known address type, check for known "untyped" formats.
	// 1.1.1.1:<port>

	lastColonIndex := strings.LastIndex(client.Addr.String(), ":")
	if lastColonIndex != -1 {
		ipString := client.Addr.String()[:lastColonIndex]
		ip := net.ParseIP(ipString)
		if ip != nil {
			return ip.String()
		}
	}

	return client.Addr.String()
}
