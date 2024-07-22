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

package kube // import "github.com/solarwinds/swi-k8s-opentelemetry-collector/processor/swk8sattributesprocessor/internal/kube"

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

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

// Add this new function for StatefulSets
func newStatefulSetSharedInformer(
	client kubernetes.Interface,
	namespace string,
	ls labels.Selector,
	fs fields.Selector,
) cache.SharedInformer {
	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc:  statefulSetInformerListFuncWithSelectors(client, namespace, ls, fs),
			WatchFunc: statefulSetInformerWatchFuncWithSelectors(client, namespace, ls, fs),
		},
		&appsv1.StatefulSet{},
		watchSyncPeriod,
	)
	return informer
}

func statefulSetInformerListFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.AppsV1().StatefulSets(namespace).List(context.Background(), opts)
	}

}

func statefulSetInformerWatchFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.WatchFunc {
	return func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.AppsV1().StatefulSets(namespace).Watch(context.Background(), opts)
	}
}

// Add this new function for ReplicaSets
func newReplicaSetSharedInformer(
	client kubernetes.Interface,
	namespace string,
	ls labels.Selector,
	fs fields.Selector,
) cache.SharedInformer {
	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc:  replicaSetInformerListFuncWithSelectors(client, namespace, ls, fs),
			WatchFunc: replicaSetInformerWatchFuncWithSelectors(client, namespace, ls, fs),
		},
		&appsv1.ReplicaSet{},
		watchSyncPeriod,
	)
	return informer
}

func replicaSetInformerListFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.AppsV1().ReplicaSets(namespace).List(context.Background(), opts)
	}

}

func replicaSetInformerWatchFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.WatchFunc {
	return func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.AppsV1().ReplicaSets(namespace).Watch(context.Background(), opts)
	}
}

// Add this new function for DaemonSets
func newDaemonSetSharedInformer(
	client kubernetes.Interface,
	namespace string,
	ls labels.Selector,
	fs fields.Selector,
) cache.SharedInformer {
	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc:  daemonSetInformerListFuncWithSelectors(client, namespace, ls, fs),
			WatchFunc: daemonSetInformerWatchFuncWithSelectors(client, namespace, ls, fs),
		},
		&appsv1.DaemonSet{},
		watchSyncPeriod,
	)
	return informer
}

func daemonSetInformerListFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.AppsV1().DaemonSets(namespace).List(context.Background(), opts)
	}

}

func daemonSetInformerWatchFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.WatchFunc {
	return func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.AppsV1().DaemonSets(namespace).Watch(context.Background(), opts)
	}
}

// Add this new function for Jobs
func newJobSharedInformer(
	client kubernetes.Interface,
	namespace string,
	ls labels.Selector,
	fs fields.Selector,
) cache.SharedInformer {
	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc:  jobInformerListFuncWithSelectors(client, namespace, ls, fs),
			WatchFunc: jobInformerWatchFuncWithSelectors(client, namespace, ls, fs),
		},
		&batchv1.Job{},
		watchSyncPeriod,
	)
	return informer
}

func jobInformerListFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.BatchV1().Jobs(namespace).List(context.Background(), opts)
	}

}

func jobInformerWatchFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.WatchFunc {
	return func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.BatchV1().Jobs(namespace).Watch(context.Background(), opts)
	}
}

// Add this new function for CronJob
func newCronJobSharedInformer(
	client kubernetes.Interface,
	namespace string,
	ls labels.Selector,
	fs fields.Selector,
) cache.SharedInformer {
	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc:  cronJobInformerListFuncWithSelectors(client, namespace, ls, fs),
			WatchFunc: cronJobInformerWatchFuncWithSelectors(client, namespace, ls, fs),
		},
		&batchv1.CronJob{},
		watchSyncPeriod,
	)
	return informer
}

func cronJobInformerListFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.BatchV1().CronJobs(namespace).List(context.Background(), opts)
	}
}

func cronJobInformerWatchFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.WatchFunc {
	return func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.BatchV1().CronJobs(namespace).Watch(context.Background(), opts)
	}
}

// Add this new function for Node
func newNodeSharedInformer(
	client kubernetes.Interface,
	namespace string,
	ls labels.Selector,
	fs fields.Selector,
) cache.SharedInformer {
	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc:  nodeInformerListFuncWithSelectors(client, ls, fs),
			WatchFunc: nodeInformerWatchFuncWithSelectors(client, ls, fs),
		},
		&corev1.Node{},
		watchSyncPeriod,
	)
	return informer
}

func nodeInformerListFuncWithSelectors(client kubernetes.Interface, ls labels.Selector, fs fields.Selector) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.CoreV1().Nodes().List(context.Background(), opts)
	}
}

func nodeInformerWatchFuncWithSelectors(client kubernetes.Interface, ls labels.Selector, fs fields.Selector) cache.WatchFunc {
	return func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.CoreV1().Nodes().Watch(context.Background(), opts)
	}
}

// Add this new function for Persistent volume
func newPersistentVolumeSharedInformer(
	client kubernetes.Interface,
	namespace string,
	ls labels.Selector,
	fs fields.Selector,
) cache.SharedInformer {
	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc:  persistentVolumeInformerListFuncWithSelectors(client, ls, fs),
			WatchFunc: persistentVolumeInformerWatchFuncWithSelectors(client, ls, fs),
		},
		&corev1.PersistentVolume{},
		watchSyncPeriod,
	)
	return informer
}

func persistentVolumeInformerListFuncWithSelectors(client kubernetes.Interface, ls labels.Selector, fs fields.Selector) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.CoreV1().PersistentVolumes().List(context.Background(), opts)
	}
}

func persistentVolumeInformerWatchFuncWithSelectors(client kubernetes.Interface, ls labels.Selector, fs fields.Selector) cache.WatchFunc {
	return func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.CoreV1().PersistentVolumes().Watch(context.Background(), opts)
	}
}

// Add this new function for Persistent volume
func newPersistentVolumeClaimSharedInformer(
	client kubernetes.Interface,
	namespace string,
	ls labels.Selector,
	fs fields.Selector,
) cache.SharedInformer {
	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc:  persistentVolumeClaimInformerListFuncWithSelectors(client, namespace, ls, fs),
			WatchFunc: persistentVolumeClaimInformerWatchFuncWithSelectors(client, namespace, ls, fs),
		},
		&corev1.PersistentVolumeClaim{},
		watchSyncPeriod,
	)
	return informer
}

func persistentVolumeClaimInformerListFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.CoreV1().PersistentVolumeClaims(namespace).List(context.Background(), opts)
	}
}

func persistentVolumeClaimInformerWatchFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.WatchFunc {
	return func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.CoreV1().PersistentVolumeClaims(namespace).Watch(context.Background(), opts)
	}
}

// Add this new function for Service
func newServiceSharedInformer(
	client kubernetes.Interface,
	namespace string,
	ls labels.Selector,
	fs fields.Selector,
) cache.SharedInformer {
	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc:  serviceInformerListFuncWithSelectors(client, namespace, ls, fs),
			WatchFunc: serviceInformerWatchFuncWithSelectors(client, namespace, ls, fs),
		},
		&corev1.Service{},
		watchSyncPeriod,
	)
	return informer
}

func serviceInformerListFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.CoreV1().Services(namespace).List(context.Background(), opts)
	}
}

func serviceInformerWatchFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.WatchFunc {
	return func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		return client.CoreV1().Services(namespace).Watch(context.Background(), opts)
	}
}
