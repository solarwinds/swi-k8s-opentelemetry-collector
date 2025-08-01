# Implementation Plan: NH-114421 Entity State Events Test Environment and Integration Tests

## Overview

This task involves creating a comprehensive test environment to reproduce various use cases for entity-state-events and relationship-state-events, and implementing integration tests to verify their collection. The plan includes developing a test application that can be instrumented by Beyla for HTTP communication testing, and creating robust integration tests for entity and relationship state events.

## Requirements

1. **Test Environment Setup**:
   - Create a test application with HTTP communication capabilities that can be instrumented by Beyla
   - Build the application from source code within the repository
   - Integrate with Skaffold build artifacts for local development
   - Update testResources.yaml to include the new test application

2. **Integration Tests**:
   - Create integration tests for entity-state-events collection
   - Create integration tests for relationship-state-events collection  
   - Verify events are captured by timeseries-mock-service and exposed in `entitystateevents.json`
   - Ensure tests follow existing patterns and integrate with current test infrastructure

3. **Documentation and Configuration**:
   - Update development documentation
   - Ensure proper Beyla instrumentation configuration
   - Integrate with existing Skaffold profiles and deployment configuration

## Implementation Steps

### 1. Create Test Application Source Code
**Location**: `tests/deploy/test-app/`

- **1.1** Create directory structure:
  ```
  tests/deploy/test-app/
  ├── Dockerfile
  ├── main.go (or main.py)
  ├── go.mod (if Go)
  └── requirements.txt (if Python)
  ```

- **1.2** Implement a simple HTTP server application:
  - Language choice: **Go** (preferred for Beyla instrumentation and better entity events generation)
  - Features needed:
    - HTTP server listening on configurable port (default: 8080)
    - Basic REST endpoints (`/health`, `/api/data`, `/api/users/{id}`)
    - Client capability to make HTTP requests to other services - other instances of the test-app
    - Configurable service name and basic metrics endpoint
    - Proper logging and graceful shutdown

- **1.3** Create Dockerfile:
  - Multi-stage build for efficient image size
  - Use minimal base image (alpine or distroless)
  - Expose necessary ports
  - Set appropriate user and security context

### 2. Update Skaffold Configuration
**Files**: skaffold.yaml

- **2.1** Add new build artifact for test application:
  ```yaml
  build:
    artifacts:
      - image: test-app
        context: tests/deploy/test-app
        docker:
          dockerfile: Dockerfile
  ```

- **2.2** Update existing build configuration to include the test app in appropriate profiles

### 3. Update Test Resources
**Files**: testResources.yaml

- **3.1** Add new Deployment for the test application:
  - Include proper labels and annotations for Beyla instrumentation
  - Configure service discovery annotations
  - Set up resource requests/limits
  - Add environment variables for configuration

- **3.2** Add Service resource for the test application:
  - Expose HTTP ports
  - Include proper selectors and labels
  - Add annotations for service discovery

- **3.3** Add additional test resources:
  - Create a second instance of the app for inter-service communication testing
  - Configure different service names and ports
  - Set up proper network policies if needed

### 5. Create Entity State Events Integration Tests
**Files**: `tests/integration/test_entity_state_events_comprehensive.py`

- **5.1** Extend existing test_entity_state_events_collection.py:
  - Add tests for different entity types (Pod, Service, Deployment, etc.)
  - Test entity lifecycle events (creation, updates, deletion)
  - Verify entity attribute completeness
  - Test entity state transitions

- **5.2** Create test functions:
  ```python
  def test_pod_entity_state_events()
  def test_service_entity_state_events()
  def test_deployment_entity_state_events()
  def test_container_entity_state_events()
  ```

- **5.3** Add helper functions:
  - Entity creation and cleanup utilities
  - Entity state verification functions
  - Attribute validation helpers

### 6. Create Relationship State Events Integration Tests
**Files**: `tests/integration/test_relationship_state_events_collection.py`

- **6.1** Create new test file for relationship events:
  - Test service-to-service relationships
  - Test pod-to-service relationships
  - Test HTTP communication relationships captured by Beyla
  - Verify relationship attribute completeness

- **6.2** Implement test scenarios:
  ```python
  def test_http_client_server_relationship()
  def test_pod_service_relationship()
  def test_service_dependency_relationship()
  def test_beyla_instrumented_relationships()
  ```

- **6.3** Create relationship validation helpers:
  - Relationship state parsing utilities
  - Communication pattern verification
  - HTTP trace correlation with relationships

### 7. Update Test Utilities
**Files**: test_utils.py

- **7.1** Add entity event specific utilities:
  ```python
  def get_entity_events_from_content(content)
  def get_relationship_events_from_content(content)
  def filter_events_by_entity_type(events, entity_type)
  def validate_entity_attributes(event, required_attrs)
  def validate_relationship_attributes(event, required_attrs)
  ```

- **7.2** Add test application deployment helpers:
  ```python
  def deploy_test_application(name, config)
  def cleanup_test_application(name)
  def wait_for_service_ready(service_name)
  def trigger_http_communication(source, target)
  ```

### 8. Update Integration Test Infrastructure
**Files**: integrationTestCronJob.yaml

- **8.1** Ensure CronJob includes new test files:
  - Add new test modules to test execution
  - Update test selection patterns
  - Configure appropriate test timeouts

- **8.2** Update test resource cleanup:
  - Ensure test applications are properly cleaned up
  - Add cleanup for relationship test scenarios

### 9. Update Mock Service Configuration
**Files**: configmap.yaml

- **9.1** Verify entitystateevents pipeline configuration:
  - Ensure proper filtering for entity events
  - Verify relationship events are captured
  - Check file export configuration

- **9.2** Update service configuration if needed:
  - Add additional debug endpoints
  - Enhance filtering capabilities
  - Improve event categorization

### 10. Documentation Updates
**Files**: development.md, README.md

- **10.1** Update development documentation:
  - Document new test application architecture
  - Add instructions for running entity/relationship tests
  - Document Beyla instrumentation testing procedures

- **10.2** Create or update integration test documentation:
  - Document test scenarios and expected outcomes
  - Add troubleshooting guide for entity events
  - Document relationship event patterns

### 11. Configuration and Profile Updates
**Files**: skaffold.yaml profiles section

- **11.1** Update existing profiles:
  - Ensure `beyla` profile includes test application
  - Update `auto-instrumentation` profile for test scenarios
  - Add test-specific environment variables

- **11.2** Consider creating new profile:
  - `entity-events-testing` profile for comprehensive testing
  - Enable all necessary components (Beyla, test apps, entity collection)
  - Configure appropriate log levels and debugging

## Testing

### Unit Tests
- **Helm Chart Tests**: Update existing Helm unit tests to verify new test resources are properly rendered
- **Configuration Tests**: Verify Beyla configuration includes test applications
- **Template Tests**: Ensure test resource templates are valid

### Integration Tests
- **Entity State Events**: Verify collection of entity events for all Kubernetes resource types
- **Relationship State Events**: Verify collection of relationship events from HTTP communication
- **Beyla Instrumentation**: Verify test applications are properly instrumented
- **End-to-End Scenarios**: Test complete entity lifecycle and relationship discovery

### Test Scenarios
1. **Basic Entity Events**: Pod creation, update, deletion events
2. **Service Relationships**: Service-to-service communication via HTTP
3. **Beyla Instrumentation**: HTTP traces generating relationship events  
4. **Multi-Application Communication**: Complex service mesh scenarios
5. **Event Persistence**: Verify events are properly stored and accessible via timeseries-mock-service

### Test Validation
- Verify events appear in `entitystateevents.json` endpoint
- Validate event structure and required attributes
- Confirm event timing and lifecycle accuracy
- Test event filtering and categorization
- Verify relationship event correlation with HTTP traces

This implementation plan provides a comprehensive approach to creating a robust test environment for entity and relationship state events while leveraging existing infrastructure and following established patterns in the codebase.