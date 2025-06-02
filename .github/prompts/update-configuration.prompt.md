---
mode: 'agent'
description: 'Update Helm chart configuration'
---

# Task
Read the Jira issue ${input:jira} description and if needed related Jira issues, comments or confluence documents.

# Context
- Relevant files: 
    - #file deploy/helm/events-collector-config.yaml
    - #file deploy/helm/gateway-collector-config.yaml
    - #file deploy/helm/metrics-collector-config.yaml
    - #file deploy/helm/node-collector-config.yaml
    - #file deploy/helm/templates/_common-config.tpl
    - #file deploy/helm/templates/_helpers.tpl
    - #file deploy/helm/values.yaml
    - #file deploy/helm/values.schema.json

# Definition of Done
- Helm unit tests are passing. Run tests with `helm unittest -u deploy/helm`.

