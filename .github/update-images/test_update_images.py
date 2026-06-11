"""Unit tests for the image updater. Run via `python -m pytest`."""

from __future__ import annotations

import io
from pathlib import Path

import pytest
from ruamel.yaml import YAML

from update_images import (
    Config,
    Rule,
    Update,
    _split_registry,
    bump_chart_version,
    find_image_refs,
    parse_tag,
    select_latest,
    update_chart_file,
)


# ---- _split_registry ----


@pytest.mark.parametrize(
    "repo,host,path",
    [
        ("busybox", "docker.io", "library/busybox"),
        ("alpine/k8s", "docker.io", "alpine/k8s"),
        ("solarwinds/swo-agent", "docker.io", "solarwinds/swo-agent"),
        ("docker.io/foo/bar", "docker.io", "foo/bar"),
        ("index.docker.io/foo/bar", "index.docker.io", "foo/bar"),
        ("ghcr.io/owner/pkg", "ghcr.io", "owner/pkg"),
        ("ghcr.io/owner/pkg/sub", "ghcr.io", "owner/pkg/sub"),
        ("quay.io/foo/bar", "quay.io", "foo/bar"),
        ("gcr.io/foo/bar", "gcr.io", "foo/bar"),
        ("us-docker.pkg.dev/proj/repo/img", "us-docker.pkg.dev", "proj/repo/img"),
        ("https://ghcr.io/owner/pkg", "ghcr.io", "owner/pkg"),
        ("registry.example.com:5000/foo/bar", "registry.example.com:5000", "foo/bar"),
    ],
)
def test_split_registry(repo, host, path):
    assert _split_registry(repo) == (host, path)


# ---- parse_tag ----


@pytest.mark.parametrize(
    "tag,expected_v,expected_pre",
    [
        ("1.2.3", False, False),
        ("v1.2.3", True, False),
        ("v1.2.3-rc.1", True, True),
        ("1.2.3-alpha.5", False, True),
    ],
)
def test_parse_tag_valid(tag, expected_v, expected_pre):
    parsed = parse_tag(tag)
    assert parsed is not None
    assert parsed.has_prefix_v is expected_v
    assert parsed.parsed.is_prerelease is expected_pre


@pytest.mark.parametrize("tag", ["latest", "1.2", "abc", "", "v1.2.3.4.5"])
def test_parse_tag_invalid(tag):
    assert parse_tag(tag) is None


# ---- select_latest ----


def test_select_latest_skips_prereleases_when_current_is_stable():
    current = parse_tag("v1.2.3")
    tags = ["v1.2.4", "v1.3.0-rc.1", "v1.2.5-alpha.1"]
    result = select_latest(tags, current, None)
    assert result is not None and result.raw == "v1.2.4"


def test_select_latest_allows_prereleases_when_current_is_prerelease():
    current = parse_tag("v1.2.3-rc.1")
    tags = ["v1.2.3-rc.2", "v1.2.3", "v1.2.4-rc.1"]
    result = select_latest(tags, current, None)
    # v1.2.4-rc.1 > v1.2.3 in semver because prereleases of newer versions still rank above
    assert result is not None and result.raw == "v1.2.4-rc.1"


def test_select_latest_preserves_v_prefix():
    current = parse_tag("v1.2.3")
    tags = ["1.2.4", "v1.2.4"]
    result = select_latest(tags, current, None)
    assert result is not None and result.raw == "v1.2.4"


def test_select_latest_preserves_no_prefix():
    current = parse_tag("1.2.3")
    tags = ["v1.2.4", "1.2.4"]
    result = select_latest(tags, current, None)
    assert result is not None and result.raw == "1.2.4"


def test_select_latest_returns_none_when_no_upgrade():
    current = parse_tag("v1.2.5")
    tags = ["v1.2.3", "v1.2.4", "v1.2.5"]
    assert select_latest(tags, current, None) is None


def test_select_latest_returns_none_when_no_candidates():
    current = parse_tag("v1.2.3")
    assert select_latest(["latest", "main"], current, None) is None


def test_select_latest_respects_pin_major():
    current = parse_tag("v1.2.3")
    rule = Rule(match="*", pin_major=1)
    tags = ["v1.2.4", "v2.0.0"]
    result = select_latest(tags, current, rule)
    assert result is not None and result.raw == "v1.2.4"


def test_select_latest_respects_pin_minor():
    current = parse_tag("v1.2.3")
    rule = Rule(match="*", pin_minor=2)
    tags = ["v1.2.99", "v1.3.0"]
    result = select_latest(tags, current, rule)
    assert result is not None and result.raw == "v1.2.99"


def test_select_latest_allow_prereleases_override():
    current = parse_tag("v1.2.3")
    rule = Rule(match="*", allow_prereleases=True)
    tags = ["v1.2.4-rc.1"]
    result = select_latest(tags, current, rule)
    assert result is not None and result.raw == "v1.2.4-rc.1"


# ---- bump_chart_version ----


@pytest.mark.parametrize(
    "old,new",
    [
        ("5.3.0-alpha.4", "5.3.0-alpha.5"),
        ("5.3.0-beta.1", "5.3.0-beta.2"),
        ("5.3.0-rc.7", "5.3.0-rc.8"),
        ("5.3.0", "5.3.1"),
        ("0.0.0", "0.0.1"),
    ],
)
def test_bump_chart_version(old, new):
    assert bump_chart_version(old) == new


def test_bump_chart_version_unparseable():
    assert bump_chart_version("not-a-version") == "not-a-version"


# ---- find_image_refs ----


def test_find_image_refs_walks_nested_yaml():
    data = YAML().load(io.StringIO(
        """
        image:
          repository: foo/bar
          tag: "1.2.3"
        nested:
          deeper:
            image:
              repository: baz/qux
              tag: v0.1.0
        list:
          - repository: list/one
            tag: "2.0.0"
        """
    ))
    refs = find_image_refs(data)
    repos = sorted(r.repository for r in refs)
    assert repos == ["baz/qux", "foo/bar", "list/one"]


def test_find_image_refs_ignores_dicts_missing_either_key():
    data = YAML().load(io.StringIO(
        """
        a:
          repository: only/repo
        b:
          tag: "1.0.0"
        c:
          repository: real/one
          tag: "1.0.0"
        """
    ))
    refs = find_image_refs(data)
    assert [r.repository for r in refs] == ["real/one"]


# ---- Config ----


def test_config_ignore_and_rules():
    cfg = Config(
        ignore={"grafana/beyla"},
        rules=[Rule(match="solarwinds/*", pin_major=1), Rule(match="other/exact")],
    )
    assert cfg.is_ignored("grafana/beyla")
    assert not cfg.is_ignored("solarwinds/foo")

    rule = cfg.rule_for("solarwinds/foo")
    assert rule is not None and rule.pin_major == 1

    assert cfg.rule_for("other/exact") is not None
    assert cfg.rule_for("other/different") is None


# ---- update_chart_file ----


def test_update_chart_file_bumps_appversion_and_version(tmp_path: Path):
    chart = tmp_path / "Chart.yaml"
    chart.write_text(
        "apiVersion: v2\n"
        "name: test\n"
        "version: 5.3.0-alpha.4\n"
        "appVersion: 0.152.0\n"
    )
    updates = [
        Update(path="otel.image", repository="solarwinds/solarwinds-otel-collector", old_tag="0.152.0", new_tag="0.153.0"),
    ]
    import logging
    changed = update_chart_file(chart, updates, logging.getLogger("test"))
    assert changed
    text = chart.read_text()
    assert "appVersion: 0.153.0" in text
    assert "version: 5.3.0-alpha.5" in text


def test_update_chart_file_skips_appversion_when_main_image_unchanged(tmp_path: Path):
    chart = tmp_path / "Chart.yaml"
    chart.write_text("version: 5.3.0\nappVersion: 0.152.0\n")
    updates = [Update(path="x", repository="busybox", old_tag="1.0", new_tag="1.1")]
    import logging
    changed = update_chart_file(chart, updates, logging.getLogger("test"))
    assert changed
    text = chart.read_text()
    assert "appVersion: 0.152.0" in text
    assert "version: 5.3.1" in text
