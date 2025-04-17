@echo off

REM Ensure the .tmp directory exists
if not exist .tmp mkdir .tmp

REM File to store the last Skaffold run ID
set "LAST_RUN_ID_FILE=.tmp\last_run_id"

REM Check if the current Skaffold run ID matches the last run ID
if exist "%LAST_RUN_ID_FILE%" (
    for /f "tokens=*" %%i in (%LAST_RUN_ID_FILE%) do set "LAST_RUN_ID=%%i"
    if "%LAST_RUN_ID%" == "%SKAFFOLD_RUN_ID%" (
        echo Skaffold run ID matches the last run ID. Skipping cleanup.
        exit /b 0
    )
)

REM Save the current Skaffold run ID
echo %SKAFFOLD_RUN_ID% > "%LAST_RUN_ID_FILE%"

REM Get the current Kubernetes context
for /f "tokens=*" %%i in ('kubectl config current-context') do set "current_context=%%i"

REM Check if the current context is 'docker-desktop' or 'default'
if not "%current_context%" == "docker-desktop" if not "%current_context%" == "default" (
    echo This script can only be run in the 'docker-desktop' or 'default' context. Current context is '%current_context%'. Exiting gracefully.
    exit /b 0
)

REM Patch the CRD to remove finalizers
echo Patching the CRD to remove finalizers
kubectl patch crd/opentelemetrycollectors.opentelemetry.io -p "{\"metadata\":{\"finalizers\":[]}}" --type=merge

REM Delete all resources in the test-namespace
echo Deleting all resources in the test-namespace
kubectl delete all --all -n test-namespace

REM Those resources are not deleted by the previous command
kubectl delete secrets --all -n test-namespace
kubectl delete configmaps --all -n test-namespace
kubectl delete persistentvolumeclaims --all -n test-namespace
kubectl delete serviceaccounts --all -n test-namespace
kubectl delete roles --all -n test-namespace
kubectl delete rolebindings --all -n test-namespace
kubectl delete networkpolicies --all -n test-namespace

REM Delete all CRDs from monitoring.coreos.com group
echo Deleting all CRDs from monitoring.coreos.com group
for /f "tokens=*" %%i in ('kubectl get crd -o jsonpath="{range .items[?(@.spec.group=='monitoring.coreos.com')]}{.metadata.name}{'\n'}{end}"') do kubectl delete crd %%i

REM Delete all CRDs from cert-manager.io group
echo Deleting all CRDs from cert-manager.io group
for /f "tokens=*" %%i in ('kubectl get crd -o jsonpath="{range .items[?(@.spec.group=='cert-manager.io')]}{.metadata.name}{'\n'}{end}"') do kubectl delete crd %%i

REM Delete all CRDs from acme.cert-manager.io group
echo Deleting all CRDs from acme.cert-manager.io group
for /f "tokens=*" %%i in ('kubectl get crd -o jsonpath="{range .items[?(@.spec.group=='acme.cert-manager.io')]}{.metadata.name}{'\n'}{end}"') do kubectl delete crd %%i