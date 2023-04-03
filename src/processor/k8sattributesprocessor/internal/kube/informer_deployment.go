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
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type InformerProviderDeployment func(
	client kubernetes.Interface,
	namespace string,
	labelSelector labels.Selector,
	fieldSelector fields.Selector,
) cache.SharedInformer

// Add this new function for Deployments
func newDeploymentSharedInformer(
	client kubernetes.Interface,
	namespace string,
	ls labels.Selector,
	fs fields.Selector,
) cache.SharedInformer {
	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc:  deploymentInformerListFuncWithSelectors(client, namespace, ls, fs),
			WatchFunc: deploymentInformerWatchFuncWithSelectors(client, namespace, ls, fs),
		},
		&appsv1.Deployment{},
		watchSyncPeriod,
	)
	return informer
}

func deploymentInformerListFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.AppsV1().Deployments(namespace).List(context.Background(), opts)
	}

}

func deploymentInformerWatchFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.WatchFunc {
	return func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.AppsV1().Deployments(namespace).Watch(context.Background(), opts)
	}
}
