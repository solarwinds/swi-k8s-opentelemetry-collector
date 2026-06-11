#!/usr/bin/env python3
"""Automated Docker image version updater for the SWO K8s collector Helm chart.

Walks `deploy/helm/values.yaml`, finds every `{repository, tag}` pair, queries
the appropriate registry for newer tags, and rewrites the file in place. When
the main collector image bumps, it also updates `Chart.yaml`'s appVersion and
patch-bumps the chart's own version.

The script is config-driven via `.github/update-images/config.yaml`, which
controls the ignore list and per-repository version-selection rules.

Run locally with `--dry-run` to print proposed updates without touching git.
"""

from __future__ import annotations

import argparse
import logging
import os
import re
import sys
import traceback
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Iterable, Optional
import requests
from github import Github, GithubException, InputGitTreeElement
from packaging import version as pkg_version
from ruamel.yaml import YAML

REPO_ROOT = Path(__file__).resolve().parents[2]
CONFIG_PATH = Path(__file__).resolve().parent / "config.yaml"
VALUES_PATH = REPO_ROOT / "deploy" / "helm" / "values.yaml"
CHART_PATH = REPO_ROOT / "deploy" / "helm" / "Chart.yaml"

MAIN_IMAGE_REPO = "solarwinds/solarwinds-otel-collector"
BRANCH_NAME = "update-docker-images"
HTTP_TIMEOUT = 30

VERSION_RE = re.compile(r"^(?P<prefix>v?)(?P<core>\d+\.\d+\.\d+)(?P<pre>-[\w.+-]+)?$")


# --------------------------------------------------------------------------- #
# Configuration                                                                #
# --------------------------------------------------------------------------- #


@dataclass
class Rule:
    match: str
    allow_prereleases: Optional[bool] = None
    pin_major: Optional[int] = None
    pin_minor: Optional[int] = None
    tag_prefix: Optional[str] = None

    def matches(self, repository: str) -> bool:
        if self.match.endswith("*"):
            return repository.startswith(self.match[:-1])
        return repository == self.match


@dataclass
class Config:
    ignore: set[str] = field(default_factory=set)
    rules: list[Rule] = field(default_factory=list)

    def is_ignored(self, repository: str) -> bool:
        return repository in self.ignore

    def rule_for(self, repository: str) -> Optional[Rule]:
        for rule in self.rules:
            if rule.matches(repository):
                return rule
        return None


def load_config(path: Path) -> Config:
    if not path.exists():
        return Config()
    with open(path, "r", encoding="utf-8") as f:
        data = YAML(typ="safe").load(f) or {}
    rules = [Rule(**r) for r in data.get("rules", []) or []]
    return Config(ignore=set(data.get("ignore", []) or []), rules=rules)


# --------------------------------------------------------------------------- #
# Version selection                                                            #
# --------------------------------------------------------------------------- #


@dataclass
class ParsedTag:
    raw: str
    parsed: pkg_version.Version
    has_prefix_v: bool


def parse_tag(tag: str) -> Optional[ParsedTag]:
    m = VERSION_RE.match(tag.strip())
    if not m:
        return None
    core = m.group("core") + (m.group("pre") or "")
    try:
        return ParsedTag(raw=tag, parsed=pkg_version.parse(core), has_prefix_v=bool(m.group("prefix")))
    except pkg_version.InvalidVersion:
        return None


def select_latest(
    tags: Iterable[str],
    current: ParsedTag,
    rule: Optional[Rule],
) -> Optional[ParsedTag]:
    """Pick the highest tag that satisfies the policy rules.

    Default policy: skip prereleases unless `current` is itself a prerelease.
    Always preserve the current tag's `v` prefix (otherwise registries that
    publish both `1.2.3` and `v1.2.3` flip-flop on every run).
    """
    allow_pre = rule.allow_prereleases if rule and rule.allow_prereleases is not None else current.parsed.is_prerelease
    required_prefix = rule.tag_prefix if rule and rule.tag_prefix is not None else ("v" if current.has_prefix_v else "")

    candidates: list[ParsedTag] = []
    for raw in tags:
        parsed = parse_tag(raw)
        if parsed is None:
            continue
        if required_prefix == "v" and not parsed.has_prefix_v:
            continue
        if required_prefix == "" and parsed.has_prefix_v:
            continue
        if not allow_pre and parsed.parsed.is_prerelease:
            continue
        if rule and rule.pin_major is not None and parsed.parsed.major != rule.pin_major:
            continue
        if rule and rule.pin_minor is not None and parsed.parsed.minor != rule.pin_minor:
            continue
        candidates.append(parsed)

    if not candidates:
        return None
    candidates.sort(key=lambda c: c.parsed, reverse=True)
    best = candidates[0]
    if best.parsed <= current.parsed:
        return None
    return best


# --------------------------------------------------------------------------- #
# Registry clients                                                             #
# --------------------------------------------------------------------------- #


_KNOWN_HOSTS = {"docker.io", "index.docker.io", "ghcr.io", "quay.io", "gcr.io", "us.gcr.io", "eu.gcr.io", "asia.gcr.io"}


def _split_registry(repository: str) -> tuple[str, str]:
    """Resolve `(hostname, repo_path)` from a Docker reference.

    Bare references like `solarwinds/swo-agent` or `busybox` are Docker Hub —
    `urlparse` would misread the first segment as a hostname, so we only treat
    a leading segment as a host when it contains a dot or is one of the
    well-known registry domains, mirroring the rules used by docker pull.
    """
    repo = repository.strip()
    if repo.startswith(("http://", "https://")):
        repo = repo.split("://", 1)[1]
    head, _, tail = repo.partition("/")
    if "." in head or ":" in head or head in _KNOWN_HOSTS:
        host = head.lower()
        path = tail
    else:
        host = "docker.io"
        path = repo if "/" in repo else f"library/{repo}"
    return host, path


class RegistryClient:
    """Resolves tags for a Docker repository across supported registries."""

    def __init__(self, github_token: str, logger: logging.Logger):
        self.github_token = github_token
        self.logger = logger
        self._session = requests.Session()

    def fetch_tags(self, repository: str) -> list[str]:
        host, path = _split_registry(repository)

        if host in {"docker.io", "index.docker.io"}:
            return self._docker_hub_tags(path)
        if host == "ghcr.io":
            return self._ghcr_tags(path)
        if host == "quay.io":
            return self._quay_tags(path)
        if host in {"gcr.io", "us.gcr.io", "eu.gcr.io", "asia.gcr.io"} or host.endswith(".pkg.dev"):
            return self._registry_v2_tags(host, path)
        self.logger.warning("Unknown registry hostname %r for %s; falling back to Docker Hub", host, repository)
        return self._docker_hub_tags(path)

    def _docker_hub_tags(self, repo_path: str, page_limit: int = 5) -> list[str]:
        url = f"https://hub.docker.com/v2/repositories/{repo_path}/tags"
        params = {"page_size": 100}
        tags: list[str] = []
        for _ in range(page_limit):
            try:
                resp = self._session.get(url, params=params, timeout=HTTP_TIMEOUT)
                resp.raise_for_status()
            except requests.RequestException as e:
                self.logger.error("Docker Hub tags fetch failed for %s: %s", repository, e)
                return tags
            data = resp.json()
            tags.extend(t["name"] for t in data.get("results", []) if t.get("name"))
            url = data.get("next")
            params = {}
            if not url:
                break
        self.logger.debug("Docker Hub: %d tags for %s", len(tags), repo_path)
        return tags

    def _ghcr_tags(self, repo_path: str) -> list[str]:
        parts = repo_path.split("/")
        if len(parts) < 2:
            return []
        owner, package_name = parts[0], "/".join(parts[1:])
        headers = {
            "Authorization": f"Bearer {self.github_token}",
            "Accept": "application/vnd.github+json",
            "X-GitHub-Api-Version": "2022-11-28",
        }
        for endpoint in (f"orgs/{owner}", f"users/{owner}"):
            url = f"https://api.github.com/{endpoint}/packages/container/{package_name}/versions"
            try:
                resp = self._session.get(url, headers=headers, params={"per_page": 100}, timeout=HTTP_TIMEOUT)
            except requests.RequestException as e:
                self.logger.debug("GHCR API error %s/%s: %s", endpoint, package_name, e)
                continue
            if resp.status_code == 404:
                continue
            if resp.status_code != 200:
                self.logger.debug("GHCR API %s returned %s: %s", url, resp.status_code, resp.text[:200])
                continue
            tags: list[str] = []
            for v in resp.json():
                tags.extend((v.get("metadata") or {}).get("container", {}).get("tags") or [])
            return tags
        # Fallback: GitHub releases of the same-named repo (works for many OSS projects).
        return self._github_releases_fallback(owner, parts[1])

    def _github_releases_fallback(self, owner: str, repo: str) -> list[str]:
        try:
            gh = Github(self.github_token)
            return [r.tag_name for r in gh.get_repo(f"{owner}/{repo}").get_releases()[:50]]
        except Exception as e:
            self.logger.debug("GitHub releases fallback failed for %s/%s: %s", owner, repo, e)
            return []

    def _quay_tags(self, repo_path: str) -> list[str]:
        url = f"https://quay.io/api/v1/repository/{repo_path}/tag/"
        try:
            resp = self._session.get(url, params={"limit": 100}, timeout=HTTP_TIMEOUT)
            resp.raise_for_status()
        except requests.RequestException as e:
            self.logger.error("Quay tags fetch failed for %s: %s", repo_path, e)
            return []
        return [t["name"] for t in resp.json().get("tags", []) if t.get("name")]

    def _registry_v2_tags(self, host: str, repo_path: str) -> list[str]:
        url = f"https://{host}/v2/{repo_path}/tags/list"
        try:
            resp = self._session.get(url, timeout=HTTP_TIMEOUT)
            resp.raise_for_status()
        except requests.RequestException as e:
            self.logger.error("Registry v2 tags fetch failed for %s/%s: %s", host, repo_path, e)
            return []
        return resp.json().get("tags", []) or []


# --------------------------------------------------------------------------- #
# YAML walking and updates                                                     #
# --------------------------------------------------------------------------- #


@dataclass
class ImageRef:
    path: str
    repository: str
    tag: str
    yaml_node: Any  # ruamel mapping that holds {repository, tag}


def find_image_refs(node: Any, path: str = "") -> list[ImageRef]:
    refs: list[ImageRef] = []
    if isinstance(node, dict):
        repo, tag = node.get("repository"), node.get("tag")
        if isinstance(repo, str) and repo and isinstance(tag, str):
            refs.append(ImageRef(path=path, repository=repo, tag=tag, yaml_node=node))
        for k, v in node.items():
            refs.extend(find_image_refs(v, f"{path}.{k}" if path else str(k)))
    elif isinstance(node, list):
        for i, item in enumerate(node):
            refs.extend(find_image_refs(item, f"{path}[{i}]"))
    return refs


@dataclass
class Update:
    path: str
    repository: str
    old_tag: str
    new_tag: str


def update_values_file(
    values_path: Path,
    registry: RegistryClient,
    config: Config,
    logger: logging.Logger,
    write: bool = True,
) -> list[Update]:
    yaml = YAML()
    yaml.preserve_quotes = True
    yaml.width = 4096
    yaml.indent(mapping=2, sequence=4, offset=2)

    with open(values_path, "r", encoding="utf-8") as f:
        data = yaml.load(f)

    updates: list[Update] = []
    for ref in find_image_refs(data):
        if not ref.tag or ref.tag.startswith(("<", "${")):
            logger.debug("Skipping %s with placeholder tag %r", ref.repository, ref.tag)
            continue
        if config.is_ignored(ref.repository):
            logger.info("Ignoring %s (config)", ref.repository)
            continue

        current = parse_tag(ref.tag)
        if current is None:
            logger.debug("Skipping %s: unparseable tag %r", ref.repository, ref.tag)
            continue

        logger.info("Checking %s:%s", ref.repository, ref.tag)
        tags = registry.fetch_tags(ref.repository)
        if not tags:
            logger.warning("No tags found for %s", ref.repository)
            continue

        latest = select_latest(tags, current, config.rule_for(ref.repository))
        if latest is None or latest.raw == ref.tag:
            continue

        ref.yaml_node["tag"] = latest.raw
        updates.append(Update(path=ref.path, repository=ref.repository, old_tag=ref.tag, new_tag=latest.raw))
        logger.info("  -> %s -> %s", ref.tag, latest.raw)

    if updates and write:
        with open(values_path, "w", encoding="utf-8") as f:
            yaml.dump(data, f)
    return updates


# --------------------------------------------------------------------------- #
# Chart.yaml version bumping                                                   #
# --------------------------------------------------------------------------- #


def bump_chart_version(version_str: str) -> str:
    """Patch-bump a chart version string. Preserves -alpha.N / -beta.N / -rc.N."""
    for marker in ("-alpha.", "-beta.", "-rc."):
        if marker in version_str:
            base, suffix = version_str.split(marker, 1)
            if suffix.isdigit():
                return f"{base}{marker}{int(suffix) + 1}"
    parts = version_str.split(".")
    if len(parts) >= 3:
        patch = parts[2].split("-", 1)[0]
        if patch.isdigit():
            return f"{parts[0]}.{parts[1]}.{int(patch) + 1}"
    return version_str


def update_chart_file(chart_path: Path, updates: list[Update], logger: logging.Logger) -> bool:
    if not updates or not chart_path.exists():
        return False

    with open(chart_path, "r", encoding="utf-8") as f:
        original = f.read()
    content = original

    main_update = next((u for u in updates if MAIN_IMAGE_REPO in u.repository), None)
    if main_update:
        new_app_version = main_update.new_tag.lstrip("v")
        content = re.sub(r"^appVersion:\s+.*$", f"appVersion: {new_app_version}", content, flags=re.MULTILINE)

    m = re.search(r"^version:\s+(.+)$", content, re.MULTILINE)
    if m:
        old_version = m.group(1).strip()
        new_version = bump_chart_version(old_version)
        if new_version != old_version:
            content = re.sub(r"^version:\s+.*$", f"version: {new_version}", content, flags=re.MULTILINE)
            logger.info("Chart version: %s -> %s", old_version, new_version)

    if content == original:
        return False
    with open(chart_path, "w", encoding="utf-8") as f:
        f.write(content)
    return True


# --------------------------------------------------------------------------- #
# Git / PR plumbing (PyGithub-based commits)                                   #
# --------------------------------------------------------------------------- #


def _pr_body(updates: list[Update]) -> str:
    lines = ["## Updated images", ""]
    lines.extend(f"- **{u.repository}**: `{u.old_tag}` -> `{u.new_tag}`" for u in updates)
    return "\n".join(lines)


def _commit_message(updates: list[Update]) -> str:
    body = "\n".join(f"- {u.repository}: {u.old_tag} -> {u.new_tag}" for u in updates)
    return f"chore: update docker image versions\n\n{body}\n"


def push_branch_and_pr(
    repo: Any,
    repo_owner: str,
    modified_files: list[Path],
    updates: list[Update],
    logger: logging.Logger,
) -> Optional[str]:
    """Force-create the update branch from default, commit modified files, open/refresh PR."""
    default_branch = repo.get_branch(repo.default_branch)
    try:
        ref = repo.get_git_ref(f"heads/{BRANCH_NAME}")
        ref.edit(sha=default_branch.commit.sha, force=True)
        logger.info("Reset branch %s to %s", BRANCH_NAME, default_branch.commit.sha[:7])
    except GithubException as e:
        if e.status != 404:
            raise
        ref = repo.create_git_ref(ref=f"refs/heads/{BRANCH_NAME}", sha=default_branch.commit.sha)
        logger.info("Created branch %s", BRANCH_NAME)
        ref = repo.get_git_ref(f"heads/{BRANCH_NAME}")

    base_commit = repo.get_git_commit(ref.object.sha)
    tree_elements = []
    for path in sorted({p for p in modified_files if p.exists()}):
        rel = path.resolve().relative_to(REPO_ROOT).as_posix()
        with open(path, "r", encoding="utf-8") as f:
            tree_elements.append(InputGitTreeElement(path=rel, mode="100644", type="blob", content=f.read()))
    if not tree_elements:
        logger.info("No file changes to commit")
        return None

    new_tree = repo.create_git_tree(tree_elements, base_tree=base_commit.tree)
    commit = repo.create_git_commit(_commit_message(updates), new_tree, [base_commit])
    ref.edit(sha=commit.sha)
    logger.info("Committed %s on %s", commit.sha[:7], BRANCH_NAME)

    title = "update docker image versions"
    body = _pr_body(updates)
    existing = next(iter(repo.get_pulls(state="open", head=f"{repo_owner}:{BRANCH_NAME}")), None)
    if existing:
        existing.edit(title=title, body=body)
        logger.info("Updated PR #%s", existing.number)
        return existing.html_url
    pr = repo.create_pull(title=title, body=body, head=BRANCH_NAME, base=repo.default_branch)
    logger.info("Opened PR #%s", pr.number)
    return pr.html_url


# --------------------------------------------------------------------------- #
# Entry point                                                                  #
# --------------------------------------------------------------------------- #


def setup_logging(verbose: bool) -> logging.Logger:
    logging.basicConfig(
        level=logging.DEBUG if verbose else logging.INFO,
        format="%(asctime)s %(levelname)s %(message)s",
        handlers=[logging.StreamHandler(sys.stdout)],
    )
    return logging.getLogger("update-images")


def parse_args(argv: Optional[list[str]] = None) -> argparse.Namespace:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--dry-run", action="store_true", help="Detect updates and write local files but skip git/PR steps.")
    p.add_argument("--verbose", "-v", action="store_true")
    return p.parse_args(argv)


def main() -> int:
    args = parse_args()
    logger = setup_logging(args.verbose)

    config = load_config(CONFIG_PATH)
    logger.info("Loaded config from %s (%d ignored, %d rules)", CONFIG_PATH, len(config.ignore), len(config.rules))

    github_token = os.environ.get("GITHUB_TOKEN")
    if not github_token and not args.dry_run:
        logger.error("GITHUB_TOKEN env var is required (use --dry-run to skip git operations)")
        return 1

    registry = RegistryClient(github_token or "", logger)

    try:
        updates = update_values_file(VALUES_PATH, registry, config, logger, write=not args.dry_run)
    except Exception as e:
        logger.error("values.yaml update failed: %s", e)
        logger.debug(traceback.format_exc())
        return 1

    if not updates:
        logger.info("No image updates available")
        return 0

    logger.info("Found %d updates", len(updates))

    if args.dry_run:
        logger.info("Dry run: skipping file writes, branch and PR creation")
        for u in updates:
            logger.info("  %s: %s -> %s", u.repository, u.old_tag, u.new_tag)
        return 0

    chart_changed = update_chart_file(CHART_PATH, updates, logger)

    repo_slug = os.environ.get("GITHUB_REPOSITORY", "")
    if "/" not in repo_slug:
        logger.error("GITHUB_REPOSITORY env var is malformed: %r", repo_slug)
        return 1
    repo_owner, _ = repo_slug.split("/", 1)
    repo = Github(github_token).get_repo(repo_slug)

    modified = [VALUES_PATH] + ([CHART_PATH] if chart_changed else [])
    try:
        pr_url = push_branch_and_pr(repo, repo_owner, modified, updates, logger)
    except Exception as e:
        logger.error("PR creation failed: %s", e)
        logger.debug(traceback.format_exc())
        return 1

    if pr_url:
        logger.info("PR: %s", pr_url)
    return 0


if __name__ == "__main__":
    sys.exit(main())
