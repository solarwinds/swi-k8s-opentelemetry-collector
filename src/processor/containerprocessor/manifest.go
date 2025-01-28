package containerprocessor

import "go.uber.org/zap"

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

type Status struct {
	ContainerStatuses     []statusContainer
	InitContainerStatuses []statusContainer
	Conditions            []struct {
		Timestamp string `json:"lastTransitionTime"`
	}
}

type Manifest struct {
	Kind   string `json:"kind"`
	Status Status `json:"status"`
	Spec   Spec   `json:"spec"`
}

func (m *Manifest) extractContainers(logger *zap.Logger) []Container {
	containers := make(map[string]Container)
	for _, container := range m.Spec.getContainers() {
		containers[container.Name] = container
	}

	changes := m.Status.Conditions
	t := changes[len(changes)-1].Timestamp
	result := make([]Container, 0)

	for _, container := range m.Status.getContainers() {
		if existing, exists := containers[container.Name]; exists {
			container.IsInitContainer = existing.IsInitContainer
			container.IsSidecarContainer = existing.IsSidecarContainer
		}
		container.Timestamp = t
		result = append(result, container)
	}

	logger.Info("Containers", zap.Any("containers", result))

	return result
}

func (s *Status) getContainers() []Container {
	var containers []Container
	for _, c := range s.ContainerStatuses {
		containers = append(containers, Container{
			Name:        c.Name,
			ContainerId: c.ContainerId,
			State:       getState(c.State),
		})
	}

	for _, ic := range s.InitContainerStatuses {
		containers = append(containers, Container{
			Name:        ic.Name,
			ContainerId: ic.ContainerId,
			State:       getState(ic.State),
		})
	}

	return containers
}

func (s *Spec) getContainers() []Container {
	var containers []Container
	for _, c := range s.Containers {
		containers = append(containers, Container{
			Name:               c.Name,
			IsInitContainer:    false,
			IsSidecarContainer: false,
		})
	}

	for _, ic := range s.InitContainers {
		containers = append(containers, Container{
			Name:               ic.Name,
			IsInitContainer:    true,
			IsSidecarContainer: ic.RestartPolicy == "Always",
		})
	}

	return containers
}

func getState(state map[string]interface{}) string {
	for key := range state {
		return key
	}
	return ""
}
