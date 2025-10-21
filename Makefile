.PHONY: integration-test integration-test-run integration-test-cleanup help

# Variables
TAGS_FILE := /tmp/tags.json
TEST_NAMESPACE := test-namespace
TEST_JOB_NAME := integration-test-manual
KUBECTL_TIMEOUT := 60s
POLL_INTERVAL := 3
SKAFFOLD_ENV := CLUSTER_NAME="cluster name" TEST_CLUSTER_NAMESPACE=$(TEST_NAMESPACE)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

integration-test: ## Run integration tests (build, deploy, test, cleanup)
	@echo "========================================="
	@echo "Starting Integration Test Suite"
	@echo "========================================="
	@$(MAKE) integration-test-run; \
	TEST_RESULT=$$?; \
	$(MAKE) integration-test-cleanup; \
	exit $$TEST_RESULT

integration-test-run: ## Internal target: Build, deploy and run tests
	@echo ""
	@echo "Step 1/3: Building images with Skaffold..."
	@echo "-----------------------------------------"
	$(SKAFFOLD_ENV) skaffold build --file-output=$(TAGS_FILE) -v info
	@echo ""
	@echo "Step 2/3: Deploying with Skaffold..."
	@echo "-----------------------------------------"
	$(SKAFFOLD_ENV) skaffold deploy --build-artifacts $(TAGS_FILE) -v info
	@echo ""
	@echo "Step 3/3: Running integration tests..."
	@echo "-----------------------------------------"
	@echo "Waiting for timeseries-mock-service to be ready..."
	kubectl wait --for=condition=ready --timeout=$(KUBECTL_TIMEOUT) pod -l app=timeseries-mock-service -n $(TEST_NAMESPACE)
	@echo ""
	@echo "Cleaning up any previous integration test jobs or pods..."
	@kubectl delete job $(TEST_JOB_NAME) -n $(TEST_NAMESPACE) --ignore-not-found=true >/dev/null 2>&1 || true
	@kubectl delete pod -l job-name=$(TEST_JOB_NAME) -n $(TEST_NAMESPACE) --ignore-not-found=true >/dev/null 2>&1 || true
	@echo ""
	@echo "Creating integration test job from CronJob..."
	kubectl create job --from=cronjob/integration-test $(TEST_JOB_NAME) -n $(TEST_NAMESPACE)
	@echo ""
	@echo "Waiting for test pod to be ready..."
	kubectl wait --for=condition=ready --timeout=$(KUBECTL_TIMEOUT) pod -l job-name=$(TEST_JOB_NAME) -n $(TEST_NAMESPACE)
	@echo ""
	@echo "Test pod is ready. Starting log streaming in background..."
	@LOG_PID=$$(kubectl logs -f -l job-name=$(TEST_JOB_NAME) -n $(TEST_NAMESPACE) 2>/dev/null & echo $$!); \
	echo ""; \
	echo "Waiting for tests to complete..."; \
	job_result=0; \
	while true; do \
		if kubectl wait --for=condition=complete --timeout=0 jobs/$(TEST_JOB_NAME) -n $(TEST_NAMESPACE) 2>/dev/null; then \
			job_result=0; \
			break; \
		fi; \
		if kubectl wait --for=condition=failed --timeout=0 jobs/$(TEST_JOB_NAME) -n $(TEST_NAMESPACE) 2>/dev/null; then \
			job_result=1; \
			break; \
		fi; \
		echo -n "."; \
		sleep $(POLL_INTERVAL); \
	done; \
	kill $$LOG_PID >/dev/null 2>&1 || true; \
	echo ""; \
	echo ""; \
	if [ $$job_result -eq 1 ]; then \
		echo "========================================"; \
		echo "Tests FAILED"; \
		echo "========================================"; \
		echo ""; \
		echo "Collecting pod logs for debugging..."; \
		mkdir -p pod-logs; \
		pods=$$(kubectl get pods -n $(TEST_NAMESPACE) -o=jsonpath='{.items[*].metadata.name}'); \
		for pod in $$pods; do \
			containers=$$(kubectl get pod $$pod -n $(TEST_NAMESPACE) -o=jsonpath='{.spec.containers[*].name}'); \
			for container in $$containers; do \
				echo "Saving logs for pod $$pod, container $$container..."; \
				kubectl logs -n $(TEST_NAMESPACE) $$pod -c $$container > pod-logs/$$pod-$$container.txt 2>&1 || true; \
			done; \
		done; \
		echo "Pod logs saved to pod-logs/ directory"; \
		echo ""; \
		echo "Collecting timeseries-mock-service JSON outputs..."; \
		mock_pods=$$(kubectl get pods -l app=timeseries-mock-service -n $(TEST_NAMESPACE) -o=jsonpath='{.items[*].metadata.name}'); \
		if [ -z "$$mock_pods" ]; then \
			echo "No timeseries-mock-service pods found."; \
		else \
			for mock_pod in $$mock_pods; do \
				for json_file in logs metrics events manifests traces entitystateevents; do \
					target_file=pod-logs/timeseries-mock-service-$$mock_pod-$$json_file.json; \
					echo "Saving $$json_file.json from $$mock_pod..."; \
					kubectl exec -n $(TEST_NAMESPACE) $$mock_pod -c file-provider -- cat /usr/share/nginx/html/$$json_file.json > $$target_file 2>/dev/null || \
					kubectl cp $(TEST_NAMESPACE)/$$mock_pod:/usr/share/nginx/html/$$json_file.json $$target_file >/dev/null 2>&1 || \
					echo "Failed to retrieve $$json_file.json from $$mock_pod"; \
				done; \
			done; \
		fi; \
		exit 1; \
	else \
		echo "========================================"; \
		echo "Tests SUCCEEDED"; \
		echo "========================================"; \
	fi

integration-test-cleanup: ## Cleanup Skaffold deployment
	@echo ""
	@echo "========================================="
	@echo "Cleaning up Skaffold deployment..."
	@echo "========================================="
	@$(SKAFFOLD_ENV) skaffold delete -v info || true
	@if [ -f $(TAGS_FILE) ]; then \
		echo "Removing tags file: $(TAGS_FILE)"; \
		rm -f $(TAGS_FILE); \
	fi
	@echo "Cleanup complete"
	@echo ""
