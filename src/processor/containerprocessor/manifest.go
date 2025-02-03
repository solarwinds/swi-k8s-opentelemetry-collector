package containerprocessor

type Manifest struct {
	Kind     string   `json:"kind"`
	Metadata Metadata `json:"metadata"`
	Status   Status   `json:"status"`
	Spec     Spec     `json:"spec"`
}

type Metadata struct {
	PodName     string `json:"name"`
	Namespace   string `json:"namespace"`
	Annotations struct {
		ClusterUid string `json:"swo.cloud.solarwinds.com/cluster-uid"`
	}
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
