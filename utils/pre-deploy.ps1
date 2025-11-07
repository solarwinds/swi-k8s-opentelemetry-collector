# Set namespace from TEST_CLUSTER_NAMESPACE env var or command line parameter with "test-namespace" as default
$NAMESPACE = if ($env:TEST_CLUSTER_NAMESPACE) { $env:TEST_CLUSTER_NAMESPACE } 
             elseif ($args[0]) { $args[0] } 
             else { "test-namespace" }

# Ensure the .tmp directory exists
$null = New-Item -ItemType Directory -Path ".tmp" -Force

# File to store the last Skaffold run ID
$LAST_RUN_ID_FILE = ".tmp\last_run_id"

# Check if the current Skaffold run ID matches the last run ID
if ((Test-Path $LAST_RUN_ID_FILE) -and ((Get-Content $LAST_RUN_ID_FILE -Raw).Trim() -eq $env:SKAFFOLD_RUN_ID)) {
    Write-Host "Skaffold run ID matches the last run ID. Skipping cleanup."
    exit 0
}

# Save the current Skaffold run ID
$env:SKAFFOLD_RUN_ID | Out-File -FilePath $LAST_RUN_ID_FILE -NoNewline

# Check if the current context is 'docker-desktop' or 'default'
$current_context = kubectl config current-context
if ($current_context -notin @("docker-desktop", "default")) {
    Write-Host "This script can only be run in the 'docker-desktop' or 'default' context. Current context is '$current_context'. Exiting gracefully."
    exit 0
}

# Patch the CRD to remove finalizers
Write-Host "Patching the CRD to remove finalizers"
if (kubectl get crd opentelemetrycollectors.opentelemetry.io 2>$null) {
    kubectl patch crd/opentelemetrycollectors.opentelemetry.io -p '{"metadata":{"finalizers":[]}}' --type=merge
}

# Delete all resources in the specified namespace
Write-Host "Deleting all resources in the $NAMESPACE namespace"
kubectl delete all --all -n $NAMESPACE

# Delete additional resources not covered by 'all'
$resourceTypes = @("secrets", "configmaps", "persistentvolumeclaims", "serviceaccounts", "roles", "rolebindings", "networkpolicies")
$resourceTypes | ForEach-Object { kubectl delete $_ --all -n $NAMESPACE }

# Delete CRDs by group
$crdGroups = @("monitoring.coreos.com", "cert-manager.io", "acme.cert-manager.io")
foreach ($group in $crdGroups) {
    Write-Host "Deleting all CRDs from $group group"
    kubectl get crd -o jsonpath="{range .items[?(@.spec.group=='$group')]}{.metadata.name}{'\n'}{end}" | 
        Where-Object { $_ } | 
        ForEach-Object { kubectl delete crd $_ --ignore-not-found }
}
