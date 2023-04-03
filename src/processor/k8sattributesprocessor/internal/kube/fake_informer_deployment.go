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

package kube // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/kube"

import (
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type FakeDeploymentInformer struct {
	*FakeController

	namespace     string
	labelSelector labels.Selector
	fieldSelector fields.Selector
}

func NewFakeDeploymentInformer(
	_ kubernetes.Interface,
	namespace string,
	labelSelector labels.Selector,
	fieldSelector fields.Selector,
) cache.SharedInformer {
	return &FakeDeploymentInformer{
		FakeController: &FakeController{},
		namespace:      namespace,
		labelSelector:  labelSelector,
		fieldSelector:  fieldSelector,
	}
}

func (f *FakeDeploymentInformer) AddEventHandler(handler cache.ResourceEventHandler) (cache.ResourceEventHandlerRegistration, error) {
	return f.AddEventHandlerWithResyncPeriod(handler, time.Second)
}

func (f *FakeDeploymentInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) (cache.ResourceEventHandlerRegistration, error) {
	return handler, nil
}

func (f *FakeDeploymentInformer) RemoveEventHandler(handle cache.ResourceEventHandlerRegistration) error {
	return nil
}

func (f *FakeDeploymentInformer) IsStopped() bool {
	return false
}

func (f *FakeDeploymentInformer) SetTransform(handler cache.TransformFunc) error {
	return nil
}

func (f *FakeDeploymentInformer) GetStore() cache.Store {
	return cache.NewStore(func(obj interface{}) (string, error) { return "", nil })
}

func (f *FakeDeploymentInformer) GetController() cache.Controller {
	return f.FakeController
}

func (f *FakeDeploymentInformer) SetWatchErrorHandler(cache.WatchErrorHandler) error {
	return nil
}
