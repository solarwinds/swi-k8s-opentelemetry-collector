---
mode: 'agent'
description: 'Generate or update Copilot instructions'
---

Analyze a repository comprehensively and generate or update GitHub Copilot instructions documentation to provide an accurate, up-to-date understanding of the project structure, purpose, and implementation details.

You will systematically scan the entire repository including README files, documentation, source code, configuration files, and scripts to extract comprehensive project information. Then create or update the `.github/copilot-instructions.md` file with structured documentation.

# Steps

1. **Repository Discovery**: Scan all directories and files, prioritizing:
   - README* files at all levels
   - docs/, documentation/, or similar folders
   - Source code files in primary languages
   - Configuration files (package.json, requirements.txt, Dockerfile, etc.)
   - Build scripts and CI/CD configurations
   - Infrastructure as Code files (Terraform, Helm, K8s manifests)

2. **Information Extraction**: For each required section, gather factual details:
   - Repository purpose: Extract from README, package descriptions, and main documentation
   - Technology stack: Identify from import statements, dependencies, configuration files
   - Directory structure: Map actual folder hierarchy and contents
   - Architecture: Analyze module imports, service connections, and component relationships
   - Build instructions: Locate scripts, Makefiles, CI configs, and deployment manifests
   - Domain configurations: Identify specialized tooling and configuration patterns
   - Conventions: Extract from existing documentation, CI rules, and code patterns

3. **Documentation Generation**: Create structured markdown following the specified format and update rules

Make sure to pay extra attention to the following sections and update them based on the current state of the repository:
* `Directory layout` section - compare information with current state of the repository and make updates if needed.
* `OTEL Collector Pipelines` section - compare information with current state of the repository (#file doc/collectorPipeline.md and actual configurations) and make updates if needed.

# Output Format

Generate a complete markdown document with the following structure. Include a JSON summary at the beginning for structured data extraction:

```json
{
  "repository_analysis": {
    "primary_languages": ["language1", "language2"],
    "frameworks": ["framework1", "framework2"],
    "key_directories": {"dir_name": "description"},
    "build_systems": ["system1", "system2"],
    "deployment_methods": ["method1", "method2"]
  }
}
```

Follow with the complete markdown document using level-2 headings (##) for each required section.

# Examples

**Example Repository Analysis JSON:**
```json
{
  "repository_analysis": {
    "primary_languages": ["TypeScript", "JavaScript", "Python"],
    "frameworks": ["React", "Node.js", "Express"],
    "key_directories": {
      "src/": "Main application source code",
      "tests/": "Unit and integration tests",
      "docs/": "Project documentation"
    },
    "build_systems": ["npm", "webpack", "Docker"],
    "deployment_methods": ["Kubernetes", "GitHub Actions"]
  }
}
```

**Example Repository Purpose Section:**
```markdown
## Repository purpose

[Project Name] is a [brief description of main functionality] designed for [target users/use cases]. It provides [key capabilities] through [main interfaces/APIs], enabling [primary business value]. The project serves [specific user groups] who need [core problems solved].
```

**Example Technology Stack Section:**
```markdown
## Technology stack

**Languages:** TypeScript (primary), Python (data processing), Shell (automation)
**Frontend:** React 18, Material-UI, Chart.js
**Backend:** Node.js, Express, PostgreSQL
**Infrastructure:** Docker, Kubernetes, AWS ECS
**CI/CD:** GitHub Actions, AWS CodePipeline
**Monitoring:** Prometheus, Grafana, CloudWatch
```

(Real examples should include all technologies actually detected in the repository with specific versions where available)

# Notes

- Only include information that actually exists in the repository - do not invent or assume details
- When updating existing documentation, preserve unchanged sections and custom content added by maintainers
- Remove outdated sections that no longer reflect the current repository state
- Focus on facts over opinions - describe what the code does, not how well it does it
- Prioritize information that helps developers understand the codebase structure and contribution workflow