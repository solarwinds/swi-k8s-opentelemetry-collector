# Metrics exported by swi-k8s-opentelemetry-collector

The following tables contain the list of all metrics exported by the swi-k8s-opentelemetry-collector. The "native" metrics are forwarded from the cluster, the "custom" metrics are calculated by the collector.

## Cluster metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.cluster.cpu.allocatable | Gauge | cores | The allocatable of CPU on cluster that are available for scheduling | custom |
| k8s.cluster.cpu.capacity | Gauge | cores | The cluster CPU capacity | custom |
| k8s.cluster.memory.allocatable | Gauge | bytes | The allocatable of memory on cluster that are available for scheduling | custom |
| k8s.cluster.memory.capacity | Gauge | bytes | The cluster memory capacity | custom |
| k8s.cluster.nodes | Gauge |  | The count of nodes on cluster | custom |
| k8s.cluster.nodes.ready | Gauge |  | The count of nodes with status condition ready | custom |
| k8s.cluster.nodes.ready.avg | Gauge | percent | The percentage of nodes with status condition ready | custom |
| k8s.cluster.pods | Gauge |  | The count of pods on cluster | custom |
| k8s.cluster.pods.running | Gauge |  | The count of pods in running phase | custom |
| k8s.cluster.spec.cpu.requests | Gauge | cores | The total number of requested CPU by all containers in a cluster | custom |
| k8s.cluster.spec.memory.requests | Gauge | bytes | The total number of requested memory by all containers in a cluster | custom |
| k8s.cluster.version | Gauge |  | Kubernetes cluster version| custom |

## Node metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.kube_node_created | Gauge | seconds | Unix creation timestamp | native |
| k8s.kube_node_info | Gauge |  | Information about a cluster node | native |
| k8s.kube_node_spec_unschedulable | Gauge |  | Whether a node can schedule new pods | native |
| k8s.kube_node_status_allocatable | Gauge |  | The amount of resources allocatable for pods (after reserving some for system daemons) | native |
| k8s.kube_node_status_capacity | Gauge | cpu=\<cores\><br>ephemeral_storage=\<bytes\><br>pods=\<integer\><br>attachable_volumes_\*=\<bytes\><br>hugepages\_\*=\<bytes\><br>memory=\<bytes\> | The total amount of resources available for a node | native |
| k8s.kube_node_status_condition | Gauge |  | The condition of a cluster node | native |
| k8s.kube_node_status_ready | Gauge |  | Node status (as tag sw.k8s.node.status) | custom |
| k8s.node.cpu.allocatable | Gauge |  | The allocatable of CPU on node that are available for scheduling | custom |
| k8s.node.cpu.capacity | Gauge |  | The node CPU capacity | custom |
| k8s.node.cpu.usage.seconds.rate | Gauge |  | The rate of node cumulative cpu time consumed | custom |
| k8s.node.fs.iops | Gauge |  | Rate of reads and writes of all pods on node | custom |
| k8s.node.fs.throughput | Gauge |  | Rate of bytes read and written of all pods on node | custom |
| k8s.node.fs.usage | Gauge |  | Number of bytes that are consumed by containers on this node’s filesystem | custom |
| k8s.node.memory.allocatable | Gauge |  | The allocatable of memory on node that are available for scheduling | custom |
| k8s.node.memory.capacity | Gauge |  | The node memory capacity | custom |
| k8s.node.memory.working_set | Gauge |  | Current working set on node | custom |
| k8s.node.network.bytes_received | Gauge |  | Rate of bytes received on node | custom |
| k8s.node.network.bytes_transmitted | Gauge |  | Rate of bytes transmitted on node | custom |
| k8s.node.network.packets_received | Gauge |  | Rate  of packets received on node | custom |
| k8s.node.network.packets_transmitted | Gauge |  | Rate of packets transmitted on node | custom |
| k8s.node.network.receive_packets_dropped | Gauge |  | Rate of packets dropped while receiving on node | custom |
| k8s.node.network.transmit_packets_dropped | Gauge |  | Rate of packets dropped while transmitting on node | custom |
| k8s.node.pods | Gauge |  | The count of pods on node | custom |
| k8s.node.status.condition.diskpressure | Gauge |  | The condition diskpressure of a cluster node (1 when true, 0 when false or unknown) | custom |
| k8s.node.status.condition.memorypressure | Gauge |  | The condition memorypressure of a cluster node (1 when true, 0 when false or unknown) | custom |
| k8s.node.status.condition.networkunavailable | Gauge |  | The condition networkunavailable of a cluster node (1 when true, 0 when false or unknown) | custom |
| k8s.node.status.condition.pidpressure | Gauge |  | The condition pidpressure of a cluster node (1 when true, 0 when false or unknown) | custom |
| k8s.node.status.condition.ready | Gauge |  | The condition ready of a cluster node (1 when true, 0 when false or unknown) | custom |

## Pod metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.kube_pod_completion_time | Gauge | seconds | Completion time in unix timestamp for a pod | native |
| k8s.kube_pod_created | Gauge | seconds | Unix creation timestamp | native |
| k8s.kube_pod_info | Gauge |  | Information about pod | native |
| k8s.kube_pod_owner | Gauge |  | Information about the Pod's owner | native |
| k8s.kube_pod_start_time | Gauge | seconds | Start time in unix timestamp for a pod | native |
| k8s.kube_pod_status_phase | Gauge |  | The pods current phase | native |
| k8s.kube_pod_status_ready | Gauge |  | Describes whether the pod is ready to serve requests | native |
| k8s.kube_pod_status_reason | Gauge |  | The pod status reasons | native |
| k8s.kube.pod.owner.daemonset | Gauge |  | Information about the DaemonSet owning the Pod | custom |
| k8s.kube.pod.owner.replicaset | Gauge |  | Information about the ReplicaSet owning the Pod | custom |
| k8s.kube.pod.owner.statefulset | Gauge |  | Information about the StatefulSet owning the Pod | custom |
| k8s.kube.pod.owner.job | Gauge |  | Information about the Job owning the Pod | custom |
| k8s.pod.containers | Gauge |  | The count of containers on pod | custom |
| k8s.pod.containers.running | Gauge |  | Current number of running containers on pod | custom |
| k8s.pod.cpu.usage.seconds.rate | Gauge | seconds | The rate of pod cumulative CPU time consumed | custom |
| k8s.pod.fs.iops | Gauge |  | Rate of reads and writes of all containers on pod | custom |
| k8s.pod.fs.reads.bytes.rate | Gauge |  | Rate of bytes read of all containers on pod | custom |
| k8s.pod.fs.reads.rate | Gauge |  | Rate of reads of all containers on pod | custom |
| k8s.pod.fs.throughput | Gauge |  | Rate of bytes read and written of all containers on pod | custom |
| k8s.pod.fs.usage.bytes | Gauge | bytes | Number of bytes that are consumed by containers on this pod's filesystem | custom |
| k8s.pod.fs.writes.bytes.rate | Gauge |  | Rate of bytes written of all containers on pod | custom |
| k8s.pod.fs.writes.rate | Gauge |  | Rate of writes of all containers on pod | custom |
| k8s.pod.memory.working_set | Gauge | bytes | Current working set on pod | custom |
| k8s.pod.network.bytes_received | Gauge |  | Rate of bytes received of all containers on pod | custom |
| k8s.pod.network.bytes_transmitted | Gauge |  | Rate of bytes transmitted of all containers on pod | custom |
| k8s.pod.network.packets_received | Gauge |  | Rate  of packets received of all containers on pod | custom |
| k8s.pod.network.packets_transmitted | Gauge |  | Rate of packets transmitted of all containers on pod | custom |
| k8s.pod.network.receive_packets_dropped | Gauge |  | Rate of packets dropped while receiving of all containers on pod | custom |
| k8s.pod.network.transmit_packets_dropped | Gauge |  | Rate of packets dropped while transmitting of all containers on pod | custom |
| k8s.pod.spec.cpu.limit | Gauge | cores | CPU quota of all containers on pod in given CPU period | custom |
| k8s.pod.spec.cpu.requests | Gauge | cores | The number of requested request resource by all containers on pod | custom |
| k8s.pod.spec.memory.limit | Gauge | bytes | Memory limit for all containers on pod | custom |
| k8s.pod.spec.memory.requests | Gauge | bytes | The number of requested memory by all containers on pod | custom |
| k8s.pod.status.reason | Gauge |  | The current pod status reason | custom |

## Container metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.container_cpu_cfs_periods_total | Counter |  | Number of elapsed enforcement period intervals | native |
| k8s.container_cpu_cfs_throttled_periods_total | Counter |  | Number of throttled period intervals | native |
| k8s.container_cpu_usage_seconds_total | Counter |  | secondsCumulative CPU time consumed | native |
| k8s.container_fs_reads_bytes_total | Counter |  | bytesCumulative count of bytes read | native |
| k8s.container_fs_reads_total | Counter |  | Cumulative count of reads completed | native |
| k8s.container_fs_usage_bytes | Gauge | bytes | Number of bytes that are consumed by the container on this filesystem | native |
| k8s.container_fs_writes_bytes_total | Counter | bytes | Cumulative count of bytes written | native |
| k8s.container_fs_writes_total | Counter |  | Cumulative count of writes completed | native |
| k8s.container_memory_working_set_bytes | Gauge | bytes | Current working set | native |
| k8s.container_network_receive_bytes_total | Counter | bytes | Cumulative count of bytes received | native |
| k8s.container_network_receive_packets_dropped_total | Counter |  | Cumulative count of packets dropped while receiving | native |
| k8s.container_network_receive_packets_total | Counter |  | Cumulative count of packets received | native |
| k8s.container_network_transmit_bytes_total | Counter | bytes | Cumulative count of bytes transmitted | native |
| k8s.container_network_transmit_packets_dropped_total | Counter |  | Cumulative count of packets dropped while transmitting | native |
| k8s.container_network_transmit_packets_total | Counter |  | Cumulative count of packets transmitted | native |
| k8s.container_spec_cpu_period | Gauge |  | CPU period of the container | native |
| k8s.container_spec_cpu_quota | Gauge |  | CPU quota of the container | native |
| k8s.container_spec_memory_limit_bytes | Gauge | bytes | Memory limit for the container | native |
| k8s.container.spec.cpu.requests | Gauge | cores | The number of requested CPU by a container | custom |
| k8s.container.spec.cpu.limit | Gauge | cores | CPU quota of container in given CPU period | custom |
| k8s.container.cpu.usage.seconds.rate | Gauge | cores | The rate of pod cumulative CPU time consumed | custom |
| k8s.container.spec.memory.requests | Gauge | bytes | The number of requested memory by a container | custom |
| k8s.container.status | Gauge |  | Describes the status of the container (waiting/running/terminated) | custom |
| k8s.container.fs.iops | Gauge |  | Rate of reads and writes on container | custom |
| k8s.container.fs.throughput | Gauge |  | Rate of bytes read and written on container | custom |
| k8s.container.network.bytes_received | Gauge |  | Rate of bytes received on container | custom |
| k8s.container.network.bytes_transmitted | Gauge |  | Rate of bytes transmitted on container | custom |
| k8s.kube_pod_container_info | Gauge |  | Information about a container in a pod | native |
| k8s.kube_pod_container_resource_limits | Gauge | cpu=\<cores\><br>memory=\<bytes\> | The number of requested limit resource by a container | native |
| k8s.kube_pod_container_resource_requests | Gauge | cpu=\<cores\><br>memory=\<bytes\> | The number of requested request resource by a container | native |
| k8s.kube_pod_container_state_started | Gauge | seconds | Start time in unix timestamp for a pod container | native |
| k8s.kube_pod_container_status_last_terminated_exitcode | Gauge |  | Describes the exit code for the last container in terminated state | native |
| k8s.kube_pod_container_status_last_terminated_reason | Gauge  | | Describes the last reason the container was in terminated state | native |
| k8s.kube_pod_container_status_ready | Gauge |  | Describes whether the containers readiness check succeeded | native |
| k8s.kube_pod_container_status_restarts_total | Counter |  | The number of container restarts per container | native |
| k8s.kube_pod_container_status_running | Gauge |  | Describes whether the container is currently in running state | native |
| k8s.kube_pod_container_status_terminated | Gauge |  | Describes whether the container is currently in terminated state | native |
| k8s.kube_pod_container_status_terminated_reason | Gauge |  | Describes the reason the container is currently in terminated state | native |
| k8s.kube_pod_container_status_waiting | Gauge |  | Describes whether the container is currently in waiting state | native |
| k8s.kube_pod_container_status_waiting_reason | Gauge |  | Describes the reason the container is currently in waiting state | native |
| k8s.kube_pod_init_container_info | Gauge |  | Information about an init container in a pod | native |
| k8s.kube_pod_init_container_status_waiting | Gauge |  | Describes whether the init container is currently in waiting state | native |
| k8s.kube_pod_init_container_status_waiting_reason | Gauge |  | Describes the reason the init container is currently in waiting state | native |
| k8s.kube_pod_init_container_status_running | Gauge |  | Describes whether the init container is currently in running state | native |
| k8s.kube_pod_init_container_status_terminated | Gauge |  | Describes whether the init container is currently in terminated state | native |
| k8s.kube_pod_init_container_status_terminated_reason | Gauge |  | Describes the reason the init container is currently in terminated state | native |
| k8s.kube_pod_init_container_status_last_terminated_reason | Gauge |  | Describes the last reason the init container was in terminated state | native |
| k8s.kube_pod_init_container_status_ready | Gauge |  | Describes whether the init containers readiness check succeeded | native |
| k8s.kube_pod_init_container_status_restarts_total | Gauge |  | The number of restarts for the init container | native |
| k8s.kube_pod_init_container_resource_limits | Gauge |  | The number of CPU cores requested limit by an init container | native |
| k8s.kube_pod_init_container_resource_requests | Gauge |  | The number of CPU cores requested by an init container | native |

## Deployment metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.deployment.condition.available | Gauge |  | Describes whether the deployment has a Available status condition | custom |
| k8s.deployment.condition.progressing | Gauge |  | Describes whether the deployment has a Progressing status condition | custom |
| k8s.deployment.condition.replicafailure | Gauge |  | Describes whether the deployment has a ReplicaFailure status condition | custom |
| k8s.kube_deployment_created | Gauge | seconds | Unix creation timestamp | native |
| k8s.kube_deployment_labels | Gauge |  | Kubernetes labels converted to Prometheus labels | native |
| k8s.kube_deployment_spec_paused | Gauge |  | Whether the deployment is paused and will not be processed by the deployment controller | native |
| k8s.kube_deployment_spec_replicas | Gauge |  | Number of desired pods for a deployment | native |
| k8s.kube_deployment_status_condition | Gauge |  | The current status conditions of a deployment | native |
| k8s.kube_deployment_status_replicas | Gauge |  | The number of replicas per deployment | native |
| k8s.kube_deployment_status_replicas_available | Gauge |  | The number of available replicas per deployment | native |
| k8s.kube_deployment_status_replicas_ready | Gauge |  | The number of ready replicas per deployment | native |
| k8s.kube_deployment_status_replicas_unavailable | Gauge |  | The number of unavailable replicas per deployment | native |
| k8s.kube_deployment_status_replicas_updated | Gauge |  | The number of updated replicas per deployment | native |

## StatefulSet metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.kube_statefulset_created | Gauge | seconds | Unix creation timestamp | native |
| k8s.kube_statefulset_labels | Gauge |  | Kubernetes labels converted to Prometheus labels | native |
| k8s.kube_statefulset_replicas | Gauge |  | Number of desired pods for a StatefulSet | native |
| k8s.kube_statefulset_status_replicas_current | Gauge |  | The number of current replicas per StatefulSet | native |
| k8s.kube_statefulset_status_replicas_ready | Gauge |  | The number of ready replicas per StatefulSet | native |
| k8s.kube_statefulset_status_replicas_updated | Gauge |  | The number of updated replicas per StatefulSet | native |

## DaemonSet metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.kube_daemonset_created | Gauge | seconds | Unix creation timestamp | native |
| k8s.kube_daemonset_labels | Gauge |  | Kubernetes labels converted to Prometheus labels | native |
| k8s.kube_daemonset_status_current_number_scheduled | Gauge |  | The number of nodes running at least one daemon pod and are supposed to | native |
| k8s.kube_daemonset_status_desired_number_scheduled | Gauge |  | The number of nodes that should be running the daemon pod | native |
| k8s.kube_daemonset_status_number_available | Gauge |  | The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and available | native |
| k8s.kube_daemonset_status_number_misscheduled | Gauge |  | The number of nodes running a daemon pod but are not supposed to | native |
| k8s.kube_daemonset_status_number_ready | Gauge |  | The number of nodes that should be running the daemon pod and have one or more of the daemon pod running and ready | native |
| k8s.kube_daemonset_status_number_unavailable | Gauge |  | The number of nodes that should be running the daemon pod and have none of the daemon pod running and available | native |
| k8s.kube_daemonset_status_updated_number_scheduled | Gauge |  | The total number of nodes that are running updated daemon pod | native |

## ReplicaSet metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.kube_replicaset_created | Gauge | seconds | Unix creation timestamp | native |
| k8s.kube_replicaset_owner | Gauge |  | Information about the ReplicaSet's owner | native |
| k8s.kube.replicaset.owner.deployment | Gauge |  | Information about the Deployment owning the ReplicaSet | custom |
| k8s.kube_replicaset_spec_replicas | Gauge |  | Information about the desired replicasets | native |
| k8s.kube_replicaset_status_ready_replicas | Gauge |  | Information about the ready replicasets | native |
| k8s.kube_replicaset_status_replicas | Gauge |  | Information about the current replicasets | native |


## Namespace metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.kube_namespace_created | Gauge | seconds | Unix creation timestamp | native |
| k8s.kube_namespace_status_phase | Gauge |  | Kubernetes namespace status phase | native |
| k8s.kube_resourcequota | Gauge |  | ResourceQuota metric | native |

## Job metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.kube_job_info | Gauge |  | Information about job | native |
| k8s.kube_job_owner | Gauge |  | Information about the Job's owner  | native |
| k8s.kube.job.owner.cronjob | Gauge |  | Information about the CronJob owning the Job | custom |
| k8s.kube_job_status_active | Gauge |  | Determine whether job is active | native |
| k8s.kube_job_status_succeeded | Gauge |  | Determine whether job succeeded | native |
| k8s.kube_job_status_failed | Gauge |  | Determine whether job failed | native |
| k8s.kube_job_status_start_time | Gauge | seconds | Unix start timestamp | native |
| k8s.kube_job_status_completion_time | Gauge | seconds | Unix completion timestamp | native |
| k8s.kube_job_complete | Gauge |  | Job completed | native |
| k8s.kube_job_failed | Gauge |  | Job failed | native |
| k8s.kube_job_created | Gauge |  seconds | Unix creation timestamp | native |
| k8s.kube_job_spec_completions | Gauge |   | Job completions | native |
| k8s.kube_job_spec_parallelism | Gauge |   | Job parallelism | native |

## Persistent Volume metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.kube_persistentvolume_capacity_bytes | Gauge |  | Information about Persistent Volume's capacity | native |
| k8s.kube_persistentvolume_info | Gauge |  | Information about Persistent Volume | native |
| k8s.kube_persistentvolume_status_phase | Gauge |  | Status of Persistent Volume | native |
| k8s.kube_persistentvolume_claim_ref | Gauge |  | Information about connected Persistent Volume Claim's | native |
| k8s.kube_persistentvolume_created | Gauge |  | Unix creation timestamp | native |
| k8s.persistentvolume.status.phase | Gauge |  | Describes the status of the Persistent Volume | custom |
| k8s.kubelet_volume_stats_available_percent | Gauge |  | The capacity in percent of the volume | native |

## Persistent Volume Claim metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.kube_persistentvolumeclaim_info | Gauge |  | Information about Persistent Volume Claim | native |
| k8s.kube_persistentvolumeclaim_access_mode | Gauge |  | Information about Persistent Volume Claim's access mode | native |
| k8s.kube_persistentvolumeclaim_status_phase | Gauge |  | Status of Persistent Volume Claim | native |
| k8s.kube_persistentvolumeclaim_resource_requests_storage_bytes | Gauge |  | Information about Persistent Volume Claim's requested storage | native |
| k8s.kube_persistentvolumeclaim_created | Gauge |  | Unix creation timestamp | native |
| k8s.kube_pod_spec_volumes_persistentvolumeclaims_info | Gauge |  | Information about which Pod is assigned to which Persistent Volume Claim | native |
| k8s.persistentvolumeclaim.status.phase | Gauge |  | Determine whether job succeeded | native |

## Service metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.kube_service_annotations | Gauge |  | Kubernetes annotations converted to Prometheus labels | native |
| k8s.kube_service_info | Gauge |  | Information about service | native |
| k8s.kube_service_labels| Gauge |  | Kubernetes labels converted to Prometheus labels | native |
| k8s.kube_service_created| Gauge |  | Unix creation timestamp | native |
| k8s.kube_service_spec_type| Gauge |  | Type about service | native |
| k8s.kube_service_spec_external_ip| Gauge |  | Service external ips. One series for each ip | native |
| k8s.kube_service_status_load_balancer_ingress | Gauge |  | Service load balancer ingress status | native |

## Endpoint metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.kube_endpoint_annotations| Gauge |  | Kubernetes annotations converted to Prometheus labels | native |
| k8s.kube_endpoint_address_not_ready| Gauge |  |  | native |
| k8s.kube_endpoint_address_available| Gauge |  |  | native |
| k8s.kube_endpoint_info| Gauge |  | Information about Endpoint | native |
| k8s.kube_endpoint_labels| Gauge |  | Kubernetes labels converted to Prometheus labels | native |
| k8s.kube_endpoint_created| Gauge |  | Unix creation timestamp | native |
| k8s.kube_endpoint_ports| Gauge |  | Endpoint port one for each series. | native |
| k8s.kube_endpoint_address| Gauge |  | Endpoint address one for each series | native |


## Other metrics

| Metric | Type | Unit | Description | native/custom |
| ---    | ---  | ---  | ---         | ---           |
| k8s.apiserver.request.successrate | Gauge | percent | Success rate of Kubernetes API server calls | custom |
