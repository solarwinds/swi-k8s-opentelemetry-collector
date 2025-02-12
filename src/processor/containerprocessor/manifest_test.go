// Copyright 2025 SolarWinds Worldwide, LLC. All rights reserved.
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

package containerprocessor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	manifest = Manifest{
		Metadata: Metadata{
			PodName:   "test-pod",
			Namespace: "test-namespace",
		},
		Status: Status{
			ContainerStatuses: []statusContainer{
				{
					Name:        "test-container",
					ContainerId: "test-container-id",
					State: map[string]interface{}{
						"running": map[string]interface{}{
							"startedAt": "2021-01-01T00:00:00Z",
						},
					},
				},
				{
					Name:        "test-container-missing-in-spec",
					ContainerId: "test-container-missing-in-spec-id",
					State: map[string]interface{}{
						"running": map[string]interface{}{
							"startedAt": "2021-01-01T00:00:00Z",
						},
					},
				},
			},
			InitContainerStatuses: []statusContainer{
				{
					Name:        "test-init-container",
					ContainerId: "test-init-container-id",
					State: map[string]interface{}{
						"waiting": map[string]interface{}{},
					},
				},
				{
					Name:        "test-sidecar-container",
					ContainerId: "test-sidecar-container-id",
					State: map[string]interface{}{
						"terminated": map[string]interface{}{},
					},
				},
			},
		},
		Spec: Spec{
			Containers: []struct {
				Name string `json:"name"`
			}{
				{
					Name: "test-container",
				},
				{
					Name: "test-container-missing-in-status",
				},
			},
			InitContainers: []struct {
				Name          string `json:"name"`
				RestartPolicy string `json:"restartPolicy"`
			}{
				{
					Name:          "test-init-container",
					RestartPolicy: "Smth",
				},
				{
					Name:          "test-sidecar-container",
					RestartPolicy: "Always",
				},
			},
		},
	}
)

func TestGetContainer(t *testing.T) {
	containers := manifest.getContainers()

	// container missing in spec should not be returned in the result
	assert.Len(t, containers, 4, "Expected 4 containers")

	// Basic container
	container, ok := containers["test-container"]
	assert.Truef(t, ok, "Expected container not found")
	assert.Equal(t, "test-container", container.Name)
	expectedContainer := Container{
		Name:               "test-container",
		ContainerId:        "test-container-id",
		State:              "running",
		IsInitContainer:    false,
		IsSidecarContainer: false,
	}

	assert.Equal(t, expectedContainer, container)

	// init container
	initContainer, ok := containers["test-init-container"]
	assert.Equal(t, "test-init-container", initContainer.Name)
	expectedInitContainer := Container{
		Name:               "test-init-container",
		ContainerId:        "test-init-container-id",
		State:              "waiting",
		IsInitContainer:    true,
		IsSidecarContainer: false,
	}

	assert.Equal(t, expectedInitContainer, initContainer)

	// sidecar container
	sidecarContainer, ok := containers["test-sidecar-container"]
	assert.Equal(t, "test-sidecar-container", sidecarContainer.Name)
	expectedSidecarContainer := Container{
		Name:               "test-sidecar-container",
		ContainerId:        "test-sidecar-container-id",
		State:              "terminated",
		IsInitContainer:    true,
		IsSidecarContainer: true,
	}
	assert.Equal(t, expectedSidecarContainer, sidecarContainer)

	// container missing in status part of the manifest should be returned
	specOnlyContainer, ok := containers["test-container-missing-in-status"]
	assert.Truef(t, ok, "Expected container not found")
	assert.Equal(t, "test-container-missing-in-status", specOnlyContainer.Name)
	expectedSpecOnlyContainer := Container{
		Name:               "test-container-missing-in-status",
		IsInitContainer:    false,
		IsSidecarContainer: false,
	}

	assert.Equal(t, expectedSpecOnlyContainer, specOnlyContainer)
}
