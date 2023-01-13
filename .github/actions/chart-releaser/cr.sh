#!/usr/bin/env bash

# Copyright The Helm Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

DEFAULT_CHART_RELEASER_VERSION=v1.5.0

show_help() {
cat << EOF
Usage: $(basename "$0") <options>

    -h, --help               Display help
    -v, --version            The chart-releaser version to use (default: $DEFAULT_CHART_RELEASER_VERSION)"
        --config             The path to the chart-releaser config file
    -d, --charts-dir         The charts directory (default: charts)
    -o, --owner              The repo owner
    -r, --repo               The repo name
    -n, --install-dir        The Path to install the cr tool
    -i, --install-only       Just install the cr tool
    -s, --skip-packaging     Skip the packaging step (run your own packaging before using the releaser)
EOF
}

main() {
    local version="$DEFAULT_CHART_RELEASER_VERSION"
    local config=
    local charts_dir=charts
    local owner=
    local repo=
    local install_dir=
    local install_only=
    local skip_packaging=

    parse_command_line "$@"

    : "${CR_TOKEN:?Environment variable CR_TOKEN must be set}"

    local repo_root
    repo_root=$(git rev-parse --show-toplevel)
    pushd "$repo_root" > /dev/null

    if ! [[ -n "$skip_packaging" ]]; then
        install_chart_releaser

        rm -rf .cr-release-packages
        mkdir -p .cr-release-packages

        rm -rf .cr-index
        mkdir -p .cr-index

        package_chart "deploy/helm"

        release_charts
        update_index
    else
        install_chart_releaser
        rm -rf .cr-index
        mkdir -p .cr-index
        release_charts
        update_index
    fi

    popd > /dev/null
}

parse_command_line() {
    while :; do
        case "${1:-}" in
            -h|--help)
                show_help
                exit
                ;;
            --config)
                if [[ -n "${2:-}" ]]; then
                    config="$2"
                    shift
                else
                    echo "ERROR: '--config' cannot be empty." >&2
                    show_help
                    exit 1
                fi
                ;;
            -v|--version)
                if [[ -n "${2:-}" ]]; then
                    version="$2"
                    shift
                else
                    echo "ERROR: '-v|--version' cannot be empty." >&2
                    show_help
                    exit 1
                fi
                ;;
            -d|--charts-dir)
                if [[ -n "${2:-}" ]]; then
                    charts_dir="$2"
                    shift
                else
                    echo "ERROR: '-d|--charts-dir' cannot be empty." >&2
                    show_help
                    exit 1
                fi
                ;;
            -o|--owner)
                if [[ -n "${2:-}" ]]; then
                    owner="$2"
                    shift
                else
                    echo "ERROR: '--owner' cannot be empty." >&2
                    show_help
                    exit 1
                fi
                ;;
            -r|--repo)
                if [[ -n "${2:-}" ]]; then
                    repo="$2"
                    shift
                else
                    echo "ERROR: '--repo' cannot be empty." >&2
                    show_help
                    exit 1
                fi
                ;;
            -n|--install-dir)
                if [[ -n "${2:-}" ]]; then
                    install_dir="$2"
                    shift
                fi
                ;;
            -i|--install-only)
                if [[ -n "${2:-}" ]]; then
                    install_only="$2"
                    shift
                fi
                ;;
            -s|--skip-packaging)
                if [[ -n "${2:-}" ]]; then
                    skip_packaging="$2"
                    shift
                fi
                ;;
            *)
                break
                ;;
        esac

        shift
    done

    if [[ -z "$owner" ]]; then
        echo "ERROR: '-o|--owner' is required." >&2
        show_help
        exit 1
    fi

    if [[ -z "$repo" ]]; then
        echo "ERROR: '-r|--repo' is required." >&2
        show_help
        exit 1
    fi

    if [[ -z "$install_dir" ]]; then
        local arch
        arch=$(uname -m)
        install_dir="$RUNNER_TOOL_CACHE/cr/$version/$arch"
    fi

    if [[ -n "$install_only" ]]; then
        echo "Will install cr tool and not run it..."
        install_chart_releaser
        exit 0
    fi
}

install_chart_releaser() {
    if [[ ! -d "$RUNNER_TOOL_CACHE" ]]; then
        echo "Cache directory '$RUNNER_TOOL_CACHE' does not exist" >&2
        exit 1
    fi

    if [[ ! -d "$install_dir" ]]; then
        mkdir -p "$install_dir"

        echo "Installing chart-releaser on $install_dir..."
        curl -sSLo cr.tar.gz "https://github.com/helm/chart-releaser/releases/download/$version/chart-releaser_${version#v}_linux_amd64.tar.gz"
        tar -xzf cr.tar.gz -C "$install_dir"
        rm -f cr.tar.gz
    fi

    echo 'Adding cr directory to PATH...'
    export PATH="$install_dir:$PATH"
}

filter_charts() {
    while read -r chart; do
        [[ ! -d "$chart" ]] && continue
        local file="$chart/Chart.yaml"
        if [[ -f "$file" ]]; then
            echo "$chart"
        else
           echo "WARNING: $file is missing, assuming that '$chart' is not a Helm chart. Skipping." 1>&2
        fi
    done
}

package_chart() {
    local chart="$1"

    local args=("$chart" --package-path .cr-release-packages)
    if [[ -n "$config" ]]; then
        args+=(--config "$config")
    fi

    echo "Packaging chart '$chart'..."
    cr package "${args[@]}"
}

release_charts() {
    local args=(-o "$owner" -r "$repo" -c "$(git rev-parse HEAD)")
    if [[ -n "$config" ]]; then
        args+=(--config "$config")
    fi

    echo 'Releasing charts...'
    cr upload "${args[@]}"
}

update_index() {
    local args=(-o "$owner" -r "$repo" --pr)
    if [[ -n "$config" ]]; then
        args+=(--config "$config")
    fi

    echo 'Updating charts repo index...'
    cr index "${args[@]}"
}

main "$@"