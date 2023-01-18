#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

main() {
    install_chart_releaser

    rm -rf .cr-release-packages
    mkdir -p .cr-release-packages

    rm -rf .cr-index
    mkdir -p .cr-index

    echo "Packaging chart ..."
    cr package "deploy/helm"

    echo 'Releasing chart...'
    cr upload -c "$(git rev-parse HEAD)"

    echo 'Updating chart repo index...'
    cr index --pr
}

install_chart_releaser() {
    local version="v1.5.0"
    local install_dir="$RUNNER_TOOL_CACHE/cr/$version/$(uname -m)"
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

main "$@"