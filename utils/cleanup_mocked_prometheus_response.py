# This script is supposed to simply preparation of mocked data for integraiton test, to keep there minimum set of datapoints
import re
import requests

now = 1675856675021 # to have unified timestamp in all datapoints

def extract_node(array_of_lines):
    for line in array_of_lines:
        if f'pod="{pod}' in line and f'namespace="{namespace}"' in line:
            match = re.search(r'node="([^"]+)"', line)
            node = match.group(1)
            return node
    return None

def replace_values(lines):
    result = []
    replacements = [
        ('node', 'test-node'),
        ('instance', 'test-node'),
    ]
    for line in lines:
        for replacement in replacements:
            line = replace_line(line, replacement[0], replacement[1])
        line = re.sub(r'(\d+)$', f'{now}', line) # replace datetime
        result.append(line)
    return result

def replace_line(line, attribute, new_value):
    match = re.search(f'{attribute}="([^"]+)"', line)
    if match:
        val = match.group(1)
        return line.replace(val, new_value)
    return line

pod = 'swi-k8s-otel-collector-swo-k8s-collector-metrics'
daemonset = 'swi-k8s-otel-collector-swo-k8s-collector-logs'
namespace = 'prometheus-system'
prometheushost = '<FILL HERE>'
url = f'https://{prometheushost}/federate?match%5B%5D=container_cpu_usage_seconds_total&match%5B%5D=container_spec_cpu_quota&match%5B%5D=container_spec_cpu_period&match%5B%5D=container_memory_working_set_bytes&match%5B%5D=container_spec_memory_limit_bytes&match%5B%5D=container_cpu_cfs_throttled_periods_total&match%5B%5D=container_cpu_cfs_periods_total&match%5B%5D=container_fs_reads_total&match%5B%5D=container_fs_writes_total&match%5B%5D=container_fs_reads_bytes_total&match%5B%5D=container_fs_writes_bytes_total&match%5B%5D=container_fs_usage_bytes&match%5B%5D=container_network_receive_bytes_total&match%5B%5D=container_network_transmit_bytes_total&match%5B%5D=container_network_receive_packets_total&match%5B%5D=container_network_transmit_packets_total&match%5B%5D=container_network_receive_packets_dropped_total&match%5B%5D=container_network_transmit_packets_dropped_total&match%5B%5D=kube_deployment_created&match%5B%5D=kube_daemonset_created&match%5B%5D=kube_namespace_created&match%5B%5D=kube_node_info&match%5B%5D=kube_node_created&match%5B%5D=kube_node_status_capacity&match%5B%5D=kube_node_status_condition&match%5B%5D=kube_pod_created&match%5B%5D=kube_pod_info&match%5B%5D=kube_pod_owner&match%5B%5D=kube_pod_completion_time&match%5B%5D=kube_pod_status_phase&match%5B%5D=kube_pod_status_ready&match%5B%5D=kube_pod_status_reason&match%5B%5D=kube_pod_start_time&match%5B%5D=%7B__name__%3D~%22kube_pod_container_.%2A%22%7D&match%5B%5D=kube_namespace_status_phase&match%5B%5D=kube_deployment_labels&match%5B%5D=kube_deployment_spec_replicas&match%5B%5D=kube_deployment_spec_paused&match%5B%5D=kube_deployment_status_replicas&match%5B%5D=kube_deployment_status_replicas_ready&match%5B%5D=kube_deployment_status_replicas_available&match%5B%5D=kube_deployment_status_replicas_updated&match%5B%5D=kube_deployment_status_replicas_unavailable&match%5B%5D=kube_deployment_status_condition&match%5B%5D=kube_replicaset_owner&match%5B%5D=kube_replicaset_created&match%5B%5D=kube_statefulset_labels&match%5B%5D=kube_statefulset_replicas&match%5B%5D=kube_statefulset_status_replicas_ready&match%5B%5D=kube_statefulset_status_replicas_current&match%5B%5D=kube_statefulset_status_replicas_updated&match%5B%5D=kube_statefulset_created&match%5B%5D=kube_daemonset_labels&match%5B%5D=kube_daemonset_status_current_number_scheduled&match%5B%5D=kube_daemonset_status_desired_number_scheduled&match%5B%5D=kube_daemonset_status_updated_number_scheduled&match%5B%5D=kube_daemonset_status_number_available&match%5B%5D=kube_daemonset_status_number_misscheduled&match%5B%5D=kube_daemonset_status_number_ready&match%5B%5D=kube_daemonset_status_number_unavailable&match%5B%5D=kube_resourcequota&match%5B%5D=kube_node_status_allocatable&match%5B%5D=kube_node_spec_unschedulable&match%5B%5D=apiserver_request_total'
response = requests.get(url)
if response.status_code == 200:
    print("Successfully downloaded!")
    lines = [line.decode() for line in response.content.splitlines()]
    node = extract_node(lines)
    result = [line for line in lines
              if (line.startswith("#")
                  or (('id="/kubepods/burstable"' in line or 'id="/kubepods"' in line or 'id="/"' in line) and (f'node="{node}"' in line))
                  or ((line.startswith("kube_node_")) and (f'node="{node}"' in line))
                  or ((f'pod="{pod}' in line or f'deployment="{pod}"' in line or f'replicaset="{pod}' in line or f'daemonset="{daemonset}"' in line) and f'namespace="{namespace}"' in line)
                  or ((line.startswith("kube_namespace_")) and (f'namespace="{namespace}"' in line))
                  )]

    replaced_result = replace_values(result)
    
    with open("build/docker/wiremockFiles/redirectPrometheusResponse.txt", "w") as f:
        for line in replaced_result:
            f.write(line + '\n')

