// Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package containerprocessor

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor/processortest"
	"testing"
	"time"
)

var (
	timestamp = pcommon.NewTimestampFromTime(time.Now())
)

func generateLogs() plog.Logs {
	l := plog.NewLogs()
	ls := l.ResourceLogs().AppendEmpty()
	ls.ScopeLogs().AppendEmpty()
	return l
}

func generatePodLogs(manifest string) plog.Logs {
	l := generateLogs()
	rl := l.ResourceLogs().At(0)
	rl.Resource().Attributes().PutBool("ORIGINAL_LOG", true)
	sl := rl.ScopeLogs().At(0)
	lr := sl.LogRecords().AppendEmpty()
	lr.Attributes().PutStr("k8s.object.kind", "Pod")
	lr.Body().SetStr(manifest)
	lr.SetObservedTimestamp(timestamp)

	return l
}

func generateDeploymentPodLogs() plog.Logs {
	l := generateLogs()
	rl := l.ResourceLogs().At(0)
	sl := rl.ScopeLogs().At(0)
	lr := sl.LogRecords().AppendEmpty()
	lr.Attributes().PutStr("k8s.object.kind", "Deployment")
	lr.Body().SetStr("")

	return l
}

func TestEmptyResourceLogs(t *testing.T) {
	// processor does not decorate empty Log structure
	ctx := context.Background()
	consumer := new(consumertest.LogsSink)
	processor, err := createLogsProcessor(ctx, processortest.NewNopSettings(), createDefaultConfig(), consumer)
	assert.Nil(t, err)

	err = processor.Start(ctx, componenttest.NewNopHost())
	assert.Nil(t, err)
	err = processor.ConsumeLogs(ctx, plog.NewLogs())

	assert.Nil(t, err)
	assert.Equal(t, 1, len(consumer.AllLogs()))
	l := consumer.AllLogs()[0]
	assert.Equal(t, 0, l.ResourceLogs().Len())
}

func TestEmptyLogRecords(t *testing.T) {
	// processor does not decorate empty Log Records
	ctx := context.Background()
	consumer := new(consumertest.LogsSink)
	processor, err := createLogsProcessor(ctx, processortest.NewNopSettings(), createDefaultConfig(), consumer)
	assert.Nil(t, err)

	err = processor.Start(ctx, componenttest.NewNopHost())
	assert.Nil(t, err)
	err = processor.ConsumeLogs(ctx, generateLogs())

	assert.Nil(t, err)
	assert.Equal(t, 1, len(consumer.AllLogs()))
	l := consumer.AllLogs()[0]
	assert.Equal(t, 0, l.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().Len())
}

func TestDifferentKindBody(t *testing.T) {
	// processor does not decorate Log Records with different kind than Pod
	ctx := context.Background()
	consumer := new(consumertest.LogsSink)
	processor, err := createLogsProcessor(ctx, processortest.NewNopSettings(), createDefaultConfig(), consumer)
	assert.Nil(t, err)

	err = processor.Start(ctx, componenttest.NewNopHost())
	assert.Nil(t, err)
	err = processor.ConsumeLogs(ctx, generateDeploymentPodLogs())
	assert.Nil(t, err)

	assert.Equal(t, 1, len(consumer.AllLogs()))
	lr := consumer.AllLogs()[0].ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0)
	assert.Equal(t, "", lr.Body().Str())
	assert.Equal(t, "Deployment", getStringValue(lr.Attributes(), "k8s.object.kind"))
}

func TestEmptyPodLogBody(t *testing.T) {
	ctx := context.Background()
	consumer := new(consumertest.LogsSink)
	processor, err := createLogsProcessor(ctx, processortest.NewNopSettings(), createDefaultConfig(), consumer)
	assert.Nil(t, err)

	err = processor.Start(ctx, componenttest.NewNopHost())
	assert.Nil(t, err)
	err = processor.ConsumeLogs(ctx, generatePodLogs(""))
	assert.Error(t, err)
}

func TestPodLogBody(t *testing.T) {
	// the manifest is in format that is sent in body of log record
	// some parts from original manifest that are not used by the algorithm are removed for simplicity
	body := "{\"apiVersion\":\"v1\",\"kind\":\"Pod\",\"metadata\":{\"annotations\":{\"checksum/config\":\"123456\",\"swo.cloud.solarwinds.com/cluster-uid\":\"test-cluster-uid\"},\"creationTimestamp\":\"2025-02-04T11:28:27Z\",\"generateName\":\"test-generate-name\",\"labels\":{\"some-label\":\"test-label\"},\"managedFields\":[{\"apiVersion\":\"v1\",\"fieldsType\":\"FieldsV1\",\"fieldsV1\":{\"f:metadata\":{\"f:annotations\":{\"f:swo.cloud.solarwinds.com/cluster-uid\":{}},\"f:generateName\":{},\"f:labels\":{\"f:app.kubernetes.io/instance\":{}}},\"f:spec\":{\"f:affinity\":{},\"f:containers\":{}}},\"manager\":\"test-manager\",\"operation\":\"Update\",\"time\":\"2025-02-04T11:28:27Z\"}],\"name\":\"test-pod-name\",\"namespace\":\"test-namespace\",\"ownerReferences\":[{\"apiVersion\":\"apps/v1\",\"blockOwnerDeletion\":true,\"controller\":true,\"kind\":\"DaemonSet\",\"name\":\"test\",\"uid\":\"123456789\"}],\"resourceVersion\":\"1.2.3\",\"uid\":\"123456789\"},\"spec\":{\"containers\":[{\"args\":[\"--warning\"],\"env\":[{\"name\":\"EBPF_NET_CLUSTER_NAME\",\"value\":\"cluster name\"}],\"image\":\"test-container-id\",\"imagePullPolicy\":\"IfNotPresent\",\"name\":\"test-container-name\",\"resources\":{\"requests\":{\"memory\":\"50Mi\"}},\"securityContext\":{\"privileged\":true},\"terminationMessagePath\":\"/dev/termination-log\",\"terminationMessagePolicy\":\"File\",\"volumeMounts\":[]}],\"dnsPolicy\":\"ClusterFirstWithHostNet\",\"enableServiceLinks\":true,\"hostNetwork\":true,\"hostPID\":true,\"initContainers\":[{\"command\":[\"sh\",\"-c\",\"some command;\"],\"env\":[{\"name\":\"EBPF_NET_INTAKE_HOST\",\"value\":\"test\"}],\"image\":\"test-image-container-image\",\"imagePullPolicy\":\"IfNotPresent\",\"name\":\"test-init-container-name\",\"resources\":{},\"terminationMessagePath\":\"/dev/termination-log\",\"terminationMessagePolicy\":\"File\",\"volumeMounts\":[]}],\"nodeName\":\"test-node\",\"nodeSelector\":{\"kubernetes.io/os\":\"linux\"},\"preemptionPolicy\":\"PreemptLowerPriority\",\"priority\":0,\"restartPolicy\":\"Always\",\"schedulerName\":\"test-scheduler\",\"securityContext\":{\"fsGroup\":0,\"runAsGroup\":0,\"runAsUser\":0},\"serviceAccount\":\"test-service-account\",\"serviceAccountName\":\"test-service-account-name\",\"terminationGracePeriodSeconds\":30,\"tolerations\":[{\"effect\":\"NoSchedule\",\"operator\":\"Exists\"}],\"volumes\":[{\"hostPath\":{\"path\":\"/\",\"type\":\"Directory\"},\"name\":\"host\"}]},\"status\":{\"conditions\":[{\"lastProbeTime\":null,\"lastTransitionTime\":\"2025-02-04T11:31:42Z\",\"message\":\"containers with unready status\",\"reason\":\"ContainersNotReady\",\"status\":\"False\",\"type\":\"ContainersReady\"},{\"lastProbeTime\":null,\"lastTransitionTime\":\"2025-02-04T11:28:27Z\",\"status\":\"True\",\"type\":\"PodScheduled\"}],\"containerStatuses\":[{\"containerID\":\"test-container-id\",\"image\":\"test-container-image-id\",\"imageID\":\"test-container-image-id\",\"lastState\":{\"terminated\":{\"containerID\":\"container-id\",\"exitCode\":255,\"finishedAt\":\"2025-02-04T11:30:24Z\",\"reason\":\"Error\",\"startedAt\":\"2025-02-04T11:29:10Z\"}},\"name\":\"test-container-name\",\"ready\":false,\"restartCount\":1,\"started\":false,\"state\":{\"terminated\":{\"containerID\":\"test-container-id\",\"exitCode\":255,\"finishedAt\":\"2025-02-04T11:31:41Z\",\"reason\":\"Error\",\"startedAt\":\"2025-02-04T11:30:25Z\"}}}],\"hostIP\":\"1.2.3.4\",\"hostIPs\":[{\"ip\":\"1.2.3.4\"}],\"initContainerStatuses\":[{\"containerID\":\"test-init-container-id\",\"image\":\"test-init-container-image\",\"imageID\":\"test-init-container-image-id\",\"lastState\":{},\"name\":\"test-init-container-name\",\"ready\":true,\"restartCount\":0,\"started\":false,\"state\":{\"terminated\":{\"containerID\":\"test-init-container-id\",\"exitCode\":0,\"finishedAt\":\"2025-02-04T11:29:09Z\",\"reason\":\"Completed\",\"startedAt\":\"2025-02-04T11:28:27Z\"}}}],\"phase\":\"Running\",\"podIP\":\"1.2.3.4\",\"podIPs\":[{\"ip\":\"1.2.3.4\"}],\"qosClass\":\"Burstable\",\"startTime\":\"2025-02-04T11:28:27Z\"}}"
	t.Setenv("CLUSTER_UID", "test-cluster-uid")
	l := generatePodLogs(body)
	ctx := context.Background()
	consumer := new(consumertest.LogsSink)
	processor, err := createLogsProcessor(ctx, processortest.NewNopSettings(), createDefaultConfig(), consumer)
	assert.Nil(t, err)

	err = processor.Start(ctx, componenttest.NewNopHost())
	assert.Nil(t, err)
	err = processor.ConsumeLogs(ctx, l)
	assert.Nil(t, err)

	result := consumer.AllLogs()
	assert.Len(t, result, 1)
	assert.Equal(t, result[0].ResourceLogs().Len(), 2)

	origLog := result[0].ResourceLogs().At(0)
	verifyOriginalLog(t, origLog, body)

	newLog := result[0].ResourceLogs().At(1)
	verifyNewLog(t, newLog, map[string]Container{
		"test-container-name": {
			Name:            "test-container-name",
			ContainerId:     "test-container-id",
			State:           "terminated",
			IsInitContainer: false,
		},
		"test-init-container-name": {
			Name:            "test-init-container-name",
			ContainerId:     "test-init-container-id",
			State:           "terminated",
			IsInitContainer: true,
		},
	})
}

func verifyOriginalLog(t *testing.T, origLog plog.ResourceLogs, expectedBody string) {
	assert.Equal(t, origLog.Resource().Attributes().Len(), 1)
	assert.Equal(t, true, getBoolValue(origLog.Resource().Attributes(), "ORIGINAL_LOG"))
	assert.Equal(t, origLog.ScopeLogs().Len(), 1)
	origBody := origLog.ScopeLogs().At(0).LogRecords().At(0).Body().Str()
	assert.Equal(t, origBody, expectedBody)
}

func verifyNewLog(t *testing.T, newLog plog.ResourceLogs, expectedContainers map[string]Container) {
	// resource
	assert.Equal(t, newLog.Resource().Attributes().Len(), 1)
	assert.Equal(t, "entitystateevent", getStringValue(newLog.Resource().Attributes(), "sw.k8s.log.type"))

	// scope logs
	sl := newLog.ScopeLogs().At(0)
	assert.Equal(t, newLog.ScopeLogs().Len(), 1)
	assert.Equal(t, sl.Scope().Attributes().Len(), 1)
	assert.Equal(t, true, getBoolValue(sl.Scope().Attributes(), "otel.entity.event_as_log"))

	// log records
	assert.Equal(t, sl.LogRecords().Len(), len(expectedContainers))
	for i := range sl.LogRecords().Len() {
		lr := sl.LogRecords().At(i)

		assert.Equal(t, timestamp, lr.ObservedTimestamp())

		attrs := lr.Attributes()
		assert.Equal(t, attrs.Len(), 4)
		assert.Equal(t, "", lr.Body().Str())

		eventType := getStringValue(attrs, "otel.entity.event.type")
		assert.Equal(t, "entity_state", eventType)
		entityType := getStringValue(attrs, "otel.entity.type")
		assert.Equal(t, "KubernetesContainer", entityType)

		ids := getMapValue(attrs, "otel.entity.id")
		assert.Equal(t, "test-pod-name", getStringValue(ids, "k8s.pod.name"))
		assert.Equal(t, "test-namespace", getStringValue(ids, "k8s.namespace.name"))
		assert.Equal(t, "test-cluster-uid", getStringValue(ids, "sw.k8s.cluster.uid"))
		containerName := getStringValue(ids, "k8s.container.name")
		c, exists := expectedContainers[containerName]
		assert.True(t, exists, "Container was not expected: %s", containerName)

		otherAttrs := getMapValue(attrs, "otel.entity.attributes")
		assert.Equal(t, c.ContainerId, getStringValue(otherAttrs, "container.id"))
		assert.Equal(t, c.State, getStringValue(otherAttrs, "sw.k8s.container.status"))
		assert.Equal(t, c.IsInitContainer, getBoolValue(otherAttrs, "sw.k8s.container.init"), "Unexpected value for sw.k8s.container.init attribute: %s", containerName)
		assert.Equal(t, c.IsSidecarContainer, getBoolValue(otherAttrs, "sw.k8s.container.sidecar"), "Unexpected value for sw.k8s.container.sidecar attribute: %s", containerName)
	}
}

func getStringValue(attrs pcommon.Map, key string) string {
	value, _ := attrs.Get(key)
	return value.AsString()
}

func getMapValue(attrs pcommon.Map, key string) pcommon.Map {
	value, _ := attrs.Get(key)
	return value.Map()
}

func getBoolValue(attrs pcommon.Map, key string) bool {
	value, _ := attrs.Get(key)
	return value.Bool()
}
