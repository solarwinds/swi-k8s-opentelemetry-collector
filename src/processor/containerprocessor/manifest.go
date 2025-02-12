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

type Manifest struct {
	Metadata Metadata `json:"metadata"`
	Status   Status   `json:"status"`
	Spec     Spec     `json:"spec"`
}

type Metadata struct {
	PodName   string `json:"name"`
	Namespace string `json:"namespace"`
}

type Status struct {
	ContainerStatuses     []statusContainer
	InitContainerStatuses []statusContainer
	Conditions            []Condition
}

type Condition struct {
	Timestamp string `json:"lastTransitionTime"`
}

type Spec struct {
	Containers []struct {
		Name string `json:"name"`
	} `json:"containers"`
	InitContainers []struct {
		Name          string `json:"name"`
		RestartPolicy string `json:"restartPolicy"`
	} `json:"initContainers"`
}

type statusContainer struct {
	Name        string                 `json:"name"`
	ContainerId string                 `json:"containerID"`
	State       map[string]interface{} `json:"state"`
}

type Container struct {
	Name               string
	ContainerId        string
	State              string
	IsInitContainer    bool
	IsSidecarContainer bool
}

// getContainers returns a map of containers from the manifest. Data of each container
// are merged from "spec" and "status" parts of the manifest.
func (m *Manifest) getContainers() map[string]Container {
	containers := make(map[string]Container, 0)
	for _, c := range m.Spec.Containers {
		containers[c.Name] = Container{
			Name:               c.Name,
			IsInitContainer:    false,
			IsSidecarContainer: false,
		}
	}

	for _, ic := range m.Spec.InitContainers {
		containers[ic.Name] = Container{
			Name:               ic.Name,
			IsInitContainer:    true,
			IsSidecarContainer: ic.RestartPolicy == "Always",
		}
	}

	m.Status.fillStates(containers)
	return containers
}

// fillStates fills the basic and init container states from the "status" part of the manifest.
func (s *Status) fillStates(containers map[string]Container) {
	for _, c := range s.ContainerStatuses {
		c.fillContainer(containers)
	}

	for _, ic := range s.InitContainerStatuses {
		ic.fillContainer(containers)
	}
}

// fillContainer fills the container with additional information from "status" part of manifest and
// updates the container in the containers map.
func (sc *statusContainer) fillContainer(containers map[string]Container) {
	c, ok := containers[sc.Name]
	if !ok {
		return
	}

	c.ContainerId = sc.ContainerId
	c.State = getState(sc.State)
	containers[sc.Name] = c
}

// getState parse the state of the container from the "state" part of the manifest.
// The state is the processor is looking for is the key in the map. The value of status key
// is ignored.
func getState(state map[string]interface{}) string {
	for key := range state {
		return key
	}
	return ""
}
