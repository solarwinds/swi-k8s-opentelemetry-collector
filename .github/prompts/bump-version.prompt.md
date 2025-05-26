---
mode: 'agent'
description: 'Bump version and create new PR'
---

# Task
Bump the version of the Helm chart in the `deploy/helm` directory in #file deploy/helm/Chart.yaml and create a new PR with the changes (create new branch for it if master is checked out). The version should be incremented according to semantic versioning rules.

Touch only #file deploy/helm/Chart.yaml and do not change any other files. You must ensure that you stage and commit only that one file, if there is something already staged, you stash it first, unstash it in the end of task. 

After you create the PR, you should provide a link to it.

