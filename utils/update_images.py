#!/usr/bin/env python3

import os
import re
import json
import logging
import sys
import traceback
from datetime import datetime, timezone
from typing import Dict, List, Optional, Any
from pathlib import Path
import requests
from github import Github, GithubException, InputGitTreeElement
from packaging import version
from ruamel.yaml import YAML


def setup_logging():
    """Set up structured logging with timestamps."""
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(levelname)s - %(message)s',
        handlers=[logging.StreamHandler(sys.stdout)]
    )
    return logging.getLogger(__name__)


class DockerImageUpdater:
    """Main class for updating Docker images in Helm charts."""
    
    def __init__(self, github_token: str):
        self.github_token = github_token
        self.github = Github(github_token)
        self.logger = setup_logging()
        
        self.timeout = 30
        self.branch_name = "update-docker-images"
        
        
        self.values_file_path = Path("deploy/helm/values.yaml")
        self.chart_file_path = Path("deploy/helm/Chart.yaml")
        
        # YAML configuration
        self.yaml = YAML()
        self.yaml.preserve_quotes = True
        self.yaml.width = 4096
        self.yaml.default_flow_style = False
        self.yaml.indent(mapping=2, sequence=4, offset=2)
        
        # Repository info
        repo_info = os.environ.get('GITHUB_REPOSITORY', '').split('/')
        if len(repo_info) == 2:
            self.repo_owner, self.repo_name = repo_info
            self.repo = self.github.get_repo(f"{self.repo_owner}/{self.repo_name}")
        else:
            raise ValueError("GITHUB_REPOSITORY environment variable not set properly")

    def get_docker_hub_tags(self, repository: str, limit: int = 200) -> List[str]:
        """Fetch limited tags from Docker Hub API."""
        try:
            all_tags = []
            
            if '/' not in repository:
                repo_path = f"library/{repository}"
            else:
                repo_path = repository
                
            url = f"https://hub.docker.com/v2/repositories/{repo_path}/tags"
            params = {"page_size": 100}
            
            while url and len(all_tags) < limit:
                self.logger.debug(f"Fetching Docker Hub tags for {repository}")
                response = requests.get(url, params=params, timeout=self.timeout)
                response.raise_for_status()
                
                data = response.json()
                tags = [tag['name'] for tag in data.get('results', [])]
                all_tags.extend(tags)
                
                url = data.get('next')
                params = {}  # Clear params for subsequent requests
                    
            self.logger.info(f"Found {len(all_tags)} tags for {repository}")
            return all_tags[:limit]
            
        except Exception as e:
            self.logger.error(f"Failed to fetch Docker Hub tags for {repository}: {e}")
            return []

    def get_ghcr_tags(self, repository: str) -> List[str]:
        """Fetch tags from GitHub Container Registry."""
        try:
            if repository.startswith('ghcr.io/'):
                repo_path = repository.replace('ghcr.io/', '')
            else:
                repo_path = repository
                
            parts = repo_path.split('/')
            if len(parts) < 2:
                return []
                
            owner = parts[0]
            package_name = '/'.join(parts[1:])
            
            tags = self._get_ghcr_api_tags(owner, package_name)
            if tags:
                return tags
                
            return self._get_github_release_tags(owner, parts[1])
            
        except Exception as e:
            self.logger.error(f"Failed to fetch GHCR tags for {repository}: {e}")
            return []

    def _get_ghcr_api_tags(self, owner: str, package_name: str) -> List[str]:
        """Get tags from GHCR API."""
        try:
            headers = {
                'Authorization': f'token {self.github_token}',
                'Accept': 'application/vnd.github.v3+json'
            }
            
            urls_to_try = [
                f"https://api.github.com/orgs/{owner}/packages/container/{package_name}/versions",
                f"https://api.github.com/users/{owner}/packages/container/{package_name}/versions"
            ]
            
            for url in urls_to_try:
                try:
                    response = requests.get(url, headers=headers, timeout=self.timeout)
                    if response.status_code == 200:
                        data = response.json()
                        tags = []
                        for version_info in data:
                            if version_info.get('metadata', {}).get('container', {}).get('tags'):
                                tags.extend(version_info['metadata']['container']['tags'])
                        return tags
                except Exception:
                    continue
                    
            return []
            
        except Exception as e:
            self.logger.debug(f"GHCR API failed for {owner}/{package_name}: {e}")
            return []

    def _get_github_release_tags(self, owner: str, repo_name: str) -> List[str]:
        """Get tags from GitHub releases as fallback."""
        try:
            releases_repo = self.github.get_repo(f"{owner}/{repo_name}")
            releases = releases_repo.get_releases()
            tags = [release.tag_name for release in releases[:50]]
            self.logger.info(f"Found {len(tags)} release tags for {owner}/{repo_name}")
            return tags
            
        except Exception as e:
            self.logger.debug(f"GitHub releases failed for {owner}/{repo_name}: {e}")
            return []

    def get_latest_version(self, repository: str, current_version: str = "") -> Optional[str]:
        """Get the latest semantic version for a Docker image."""
        clean_repo = repository.strip().replace('docker.io/', '').replace('index.docker.io/', '')
        
        if clean_repo.startswith('ghcr.io/'):
            tags = self.get_ghcr_tags(clean_repo)
        elif '/' in clean_repo and not clean_repo.startswith('library/'):
            # Extract registry/host from the repository string
            registry = clean_repo.split('/')[0] if '/' in clean_repo else ''
            if registry in ['ghcr.io', 'gcr.io', 'quay.io']:
                if registry == 'ghcr.io':
                    tags = self.get_ghcr_tags(clean_repo)
                else:
                    tags = self.get_docker_hub_tags(clean_repo)
            else:
                tags = self.get_docker_hub_tags(clean_repo)
        else:
            tags = self.get_docker_hub_tags(clean_repo)
            
        if not tags:
            self.logger.warning(f"No tags found for {repository}")
            return None
            
        valid_versions = []
        version_pattern = re.compile(r'^v?(\d+\.\d+\.\d+(?:-[\w\.-]+)?)$')
        
        for tag in tags:
            match = version_pattern.match(tag)
            if match:
                try:
                    if 'solarwinds-otel-collector' in repository and not tag.startswith('v'):
                        valid_versions.append((version.parse(match.group(1)), tag))
                    else:
                        valid_versions.append((version.parse(match.group(1)), tag))
                except Exception:
                    continue
                    
        if not valid_versions:
            self.logger.warning(f"No valid semantic versions found for {repository}")
            return None
            
        valid_versions.sort(key=lambda x: x[0], reverse=True)
        latest_tag = valid_versions[0][1]
        
        if current_version:
            try:
                current_parsed = version.parse(current_version.lstrip('v'))
                latest_parsed = valid_versions[0][0]
                
                if latest_parsed <= current_parsed:
                    return None
            except Exception as e:
                self.logger.debug(f"Version comparison failed for {repository}: {e}")
                
        self.logger.info(f"{repository}: Latest version {latest_tag} (current: {current_version})")
        return latest_tag

    def find_images_in_yaml(self, yaml_data: Any, path: str = "") -> List[Dict[str, Any]]:
        """Recursively find image configurations in YAML data."""
        images = []
        
        if isinstance(yaml_data, dict):
            if 'repository' in yaml_data and 'tag' in yaml_data:
                images.append({
                    'path': path,
                    'repository': yaml_data['repository'],
                    'tag': yaml_data['tag'],
                    'yaml_data': yaml_data
                })
            else:
                for key, value in yaml_data.items():
                    new_path = f"{path}.{key}" if path else key
                    images.extend(self.find_images_in_yaml(value, new_path))
                    
        elif isinstance(yaml_data, list):
            for i, item in enumerate(yaml_data):
                new_path = f"{path}[{i}]"
                images.extend(self.find_images_in_yaml(item, new_path))
                
        return images

    def update_values_yaml(self) -> List[Dict[str, Any]]:
        """Update image tags in values.yaml file."""
        if not self.values_file_path.exists():
            self.logger.error(f"Values file not found: {self.values_file_path}")
            return []
            
        with open(self.values_file_path, 'r') as f:
            yaml_data = self.yaml.load(f)
            
        images = self.find_images_in_yaml(yaml_data)
        updates = []
        
        for image_config in images:
            repository = image_config['repository']
            current_tag = image_config['tag']
            path = image_config['path']
                
            if not current_tag or current_tag.startswith('<') or current_tag.startswith('${'):
                self.logger.debug(f"Skipping {repository} with placeholder tag: {current_tag}")
                continue
                
            self.logger.info(f"Checking {repository}:{current_tag}")
            latest_tag = self.get_latest_version(repository, current_tag)
            
            if latest_tag and latest_tag != current_tag:
                image_config['yaml_data']['tag'] = latest_tag
                
                updates.append({
                    'path': path,
                    'repository': repository,
                    'old_tag': current_tag,
                    'new_tag': latest_tag
                })
                self.logger.info(f"Updated {repository}: {current_tag} → {latest_tag}")
                
        if updates:
            with open(self.values_file_path, 'w') as f:
                self.yaml.dump(yaml_data, f)
                
        return updates

    def _bump_version(self, old_version: str) -> str:
        """Bump version using semantic versioning rules."""
        try:
            if '-alpha.' in old_version:
                base_version, alpha_part = old_version.split('-alpha.', 1)
                if alpha_part.isdigit():
                    return f"{base_version}-alpha.{int(alpha_part) + 1}"
            elif '-beta.' in old_version:
                base_version, beta_part = old_version.split('-beta.', 1)
                if beta_part.isdigit():
                    return f"{base_version}-beta.{int(beta_part) + 1}"
            
            # Handle standard semantic versions
            version_parts = old_version.split('.')
            if len(version_parts) >= 3:
                patch_part = version_parts[2].split('-')[0]  # Handle pre-release suffixes
                if patch_part.isdigit():
                    new_patch = str(int(patch_part) + 1)
                    return f"{version_parts[0]}.{version_parts[1]}.{new_patch}"
                    
        except Exception as e:
            self.logger.warning(f"Could not parse version {old_version}: {e}")
            
        return old_version  # Return original if parsing fails

    def update_chart_version(self, updates: List[Dict[str, Any]]) -> bool:
        """Update Chart.yaml version and appVersion minimally."""
        if not updates or not self.chart_file_path.exists():
            return False
            
        try:
            with open(self.chart_file_path, 'r') as f:
                content = f.read()
                
            original_content = content
            
            # Find main collector image update for appVersion
            main_image_update = None
            for update in updates:
                if 'solarwinds-otel-collector' in update['repository']:
                    main_image_update = update
                    break
                    
            # Update appVersion if main image was updated
            if main_image_update:
                new_app_version = main_image_update['new_tag'].lstrip('v')
                content = re.sub(
                    r'^appVersion:\s+.*$',
                    f'appVersion: {new_app_version}',
                    content,
                    flags=re.MULTILINE
                )
                    
            # Bump chart version (patch version)
            version_match = re.search(r'^version:\s+(.+)$', content, re.MULTILINE)
            if version_match:
                old_version = version_match.group(1)
                new_version = self._bump_version(old_version)
                
                if new_version != old_version:
                    content = re.sub(
                        r'^version:\s+.*$',
                        f'version: {new_version}',
                        content,
                        flags=re.MULTILINE
                    )
                    self.logger.info(f"Updated Chart version: {old_version} → {new_version}")
                
            # Only write if content changed
            if content != original_content:
                with open(self.chart_file_path, 'w') as f:
                    f.write(content)
                    
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to update Chart.yaml: {e}")
            return False

    def create_or_update_branch(self, updates: List[Dict[str, Any]]) -> bool:
        """Create or update the update branch with changes."""
        if not updates:
            return False
            
        try:
            main_branch = self.repo.get_branch(self.repo.default_branch)
            
            try:
                branch_ref = self.repo.get_git_ref(f"heads/{self.branch_name}")
                branch_ref.edit(sha=main_branch.commit.sha, force=True)
                self.logger.info(f"Updated branch {self.branch_name}")
            except GithubException as e:
                if e.status == 404:
                    self.repo.create_git_ref(
                        ref=f"refs/heads/{self.branch_name}",
                        sha=main_branch.commit.sha
                    )
                    self.logger.info(f"Created new branch {self.branch_name}")
                else:
                    raise
                
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to create/update branch: {e}")
            return False

    def commit_changes(self, updates: List[Dict[str, Any]]) -> bool:
        """Commit changes to the update branch."""
        if not updates:
            return True
            
        try:
            commit_message = f"chore: update docker image versions\n\n"
            
            for update in updates:
                commit_message += f"- {update['repository']}: {update['old_tag']} → {update['new_tag']}\n"
                
            branch_ref = self.repo.get_git_ref(f"heads/{self.branch_name}")
            base_commit = self.repo.get_git_commit(branch_ref.object.sha)
            
            # Get current tree
            current_tree = base_commit.tree
            
            # Prepare updated files
            tree_elements = []
            
            if self.values_file_path.exists():
                with open(self.values_file_path, 'r', encoding='utf-8') as f:
                    values_content = f.read()
                tree_elements.append(InputGitTreeElement(
                    path=str(self.values_file_path),
                    mode='100644',
                    type='blob',
                    content=values_content
                ))
                
            if self.chart_file_path.exists():
                with open(self.chart_file_path, 'r', encoding='utf-8') as f:
                    chart_content = f.read()
                tree_elements.append(InputGitTreeElement(
                    path=str(self.chart_file_path),
                    mode='100644',
                    type='blob',
                    content=chart_content
                ))
                
            # Create new tree based on current tree
            new_tree = self.repo.create_git_tree(tree_elements, base_tree=current_tree)
            
            # Create commit
            commit = self.repo.create_git_commit(commit_message, new_tree, [base_commit])
            
            # Update branch reference
            branch_ref.edit(sha=commit.sha)
            
            self.logger.info(f"Committed changes to {self.branch_name}")
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to commit changes: {e}")
            self.logger.debug(traceback.format_exc())
            return False

    def create_or_update_pr(self, updates: List[Dict[str, Any]]) -> Optional[str]:
        """Create or update pull request with changes."""
        if not updates:
            return None
            
        try:
            existing_pr = None
            prs = self.repo.get_pulls(state='open', head=f"{self.repo_owner}:{self.branch_name}")
            for pr in prs:
                existing_pr = pr
                break
                
            title = f"update docker image versions"
            
            # Simple PR body
            body_parts = [
                "## Updated Images",
                ""
            ]
            
            for update in updates:
                body_parts.append(f"- **{update['repository']}**: `{update['old_tag']}` → `{update['new_tag']}`")
            
            body = "\n".join(body_parts)
            
            if existing_pr:
                existing_pr.edit(title=title, body=body)
                self.logger.info(f"Updated existing PR #{existing_pr.number}")
                return existing_pr.html_url
            else:
                new_pr = self.repo.create_pull(
                    title=title,
                    body=body,
                    head=self.branch_name,
                    base=self.repo.default_branch
                )
                self.logger.info(f"Created new PR #{new_pr.number}")
                return new_pr.html_url
                    
        except Exception as e:
            self.logger.error(f"Failed to create/update PR: {e}")
            return None

    def save_changes_log(self, updates: List[Dict[str, Any]]):
        """Save changes to JSON file for debugging."""
        log_data = {
            'timestamp': datetime.now(timezone.utc).isoformat(),
            'updates': updates,
            'summary': {
                'total_updates': len(updates),
                'repositories_updated': len(set(u['repository'] for u in updates))
            }
        }
        
        filename = f"changes_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        with open(filename, 'w') as f:
            json.dump(log_data, f, indent=2)
            
        self.logger.info(f"Saved changes log to {filename}")

    def run(self) -> bool:
        """Main execution method."""
        self.logger.info("Starting Docker Image Updater")
        
        try:
            updates = self.update_values_yaml()
            
            if not updates:
                self.logger.info("No image updates found")
                return True
                
            self.logger.info(f"Found {len(updates)} image updates")
            
            self.update_chart_version(updates)
            self.save_changes_log(updates)
                
            if not self.create_or_update_branch(updates):
                return False
                
            if not self.commit_changes(updates):
                return False
                
            pr_url = self.create_or_update_pr(updates)
            if pr_url:
                self.logger.info(f"PR available at: {pr_url}")
                
            self.logger.info("Docker Image Updater completed successfully")
            return True
            
        except Exception as e:
            self.logger.error(f"Docker Image Updater failed: {e}")
            self.logger.error(traceback.format_exc())
            return False


def main():
    """Main entry point."""
    github_token = os.environ.get('GITHUB_TOKEN')
    if not github_token:
        print("ERROR: GITHUB_TOKEN environment variable is required")
        sys.exit(1)
        
    updater = DockerImageUpdater(github_token=github_token)
    
    success = updater.run()
    sys.exit(0 if success else 1)


if __name__ == '__main__':
    main()
