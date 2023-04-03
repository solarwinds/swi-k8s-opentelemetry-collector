// Copyright 2020 OpenTelemetry Authors
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

package k8sattributesprocessor

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/kube"
	"k8s.io/client-go/tools/cache"
)

// fakeClient is used as a replacement for WatchClient in test cases.
type fakeDeploymentClient struct {
	Pods                   map[kube.PodIdentifier]*kube.Pod
	Deployments            map[kube.DeploymentIdentifier]*kube.Deployment
	Rules                  kube.ExtractionRules
	Filters                kube.Filters
	PodAssociations        []kube.Association
	DeploymentAssociations []kube.Association
	Informer               cache.SharedInformer
	NamespaceInformer      cache.SharedInformer
	Namespaces             map[string]*kube.Namespace
	StopCh                 chan struct{}
}

func (f *fakeClient) GetDeployment(identifier kube.DeploymentIdentifier) (*kube.Deployment, bool) {
	p, ok := f.Deployments[identifier]
	return p, ok
}
