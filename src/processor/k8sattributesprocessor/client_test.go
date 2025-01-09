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
// Changes customizing the original source code: see CHANGELOG.md in deploy/helm directory

package swk8sattributesprocessor

import (
	"time"

	"go.opentelemetry.io/collector/component"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"github.com/solarwinds/swi-k8s-opentelemetry-collector/internal/k8sconfig"
	"github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swk8sattributesprocessor/internal/kube"
)

// fakeClient is used as a replacement for WatchClient in test cases.
type fakeClient struct {
	Pods              map[kube.PodIdentifier]*kube.Pod
	Rules             kube.ExtractionRules
	Filters           kube.Filters
	PodAssociations   []kube.Association
	Informer          cache.SharedInformer
	NamespaceInformer cache.SharedInformer
	Namespaces        map[string]*kube.Namespace
	StopCh            chan struct{}

	Resources map[string]map[kube.ResourceIdentifier]kube.KubernetesResource
}

func selectors() (labels.Selector, fields.Selector) {
	var selectors []fields.Selector
	return labels.Everything(), fields.AndSelectors(selectors...)
}

// newFakeClient instantiates a new FakeClient object and satisfies the ClientProvider type
func newFakeClient(
	_ component.TelemetrySettings,
	apiCfg k8sconfig.APIConfig,
	rules kube.ExtractionRules,
	filters kube.Filters,
	podAssociations []kube.Association,
	exclude kube.Excludes,
	_ kube.APIClientsetProvider,
	_ kube.InformerProvider,
	_ kube.InformerProviderNamespace,
	_ bool,
	_ time.Duration,
	_ map[string]*kube.ClientResource) (kube.Client, error) {
	cs := fake.NewSimpleClientset()

	ls, fs := selectors()
	return &fakeClient{
		Pods: map[kube.PodIdentifier]*kube.Pod{},
		Resources: map[string]map[kube.ResourceIdentifier]kube.KubernetesResource{
			kube.MetadataFromDeployment:            {},
			kube.MetadataFromStatefulSet:           {},
			kube.MetadataFromReplicaSet:            {},
			kube.MetadataFromDaemonSet:             {},
			kube.MetadataFromJob:                   {},
			kube.MetadataFromCronJob:               {},
			kube.MetadataFromNode:                  {},
			kube.MetadataFromPersistentVolume:      {},
			kube.MetadataFromPersistentVolumeClaim: {},
			kube.MetadataFromService:               {},
		},
		Rules:             rules,
		Filters:           filters,
		PodAssociations:   podAssociations,
		Informer:          kube.NewFakeInformer(cs, "", ls, fs),
		NamespaceInformer: kube.NewFakeInformer(cs, "", ls, fs),
		StopCh:            make(chan struct{}),
	}, nil
}

// GetPod looks up FakeClient.Pods map by the provided string,
// which might represent either IP address or Pod UID.
func (f *fakeClient) GetPod(identifier kube.PodIdentifier) (*kube.Pod, bool) {
	p, ok := f.Pods[identifier]
	return p, ok
}

func (f *fakeClient) GetNamespace(namespace string) (*kube.Namespace, bool) {
	ns, ok := f.Namespaces[namespace]
	return ns, ok
}

// Start is a noop for FakeClient.
func (f *fakeClient) Start() error {
	if f.Informer != nil {
		go f.Informer.Run(f.StopCh)
	}
	return nil
}

// Stop is a noop for FakeClient.
func (f *fakeClient) Stop() {
	close(f.StopCh)
}
