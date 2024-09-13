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
    
    export APP_VERSION=$(git tag --sort=version:refname | grep -E '^[0-9]+.[0-9]+.[0-9]+$' | tail -n 1)
    export CHART_VERSION=$(git tag --sort=version:refname | grep -E  '^swo-k8s-collector-' | tail -n 1 | awk -F'swo-k8s-collector-' '{print $2}')
    echo "App Version=$APP_VERSION"
    echo "Chart Version=$CHART_VERSION"
    
    yq eval '.appVersion = env(APP_VERSION)' -i deploy/helm/Chart.yaml
    yq eval '.version = env(CHART_VERSION)' -i deploy/helm/Chart.yaml

    echo "Packaging chart ..."
    cr package "deploy/helm"

    echo 'Releasing chart...'
    cr upload -c "$(git rev-parse HEAD)"

    echo 'Updating chart repo index...'
    cr index

    # Find the .tgz file and extract the release name
    RELEASE_FILE=$(find .cr-release-packages -name '*.tgz')
    RELEASE_NAME=$(basename "$RELEASE_FILE" .tgz)

    echo "Release file: $RELEASE_FILE"
    echo "Release name: $RELEASE_NAME"

    echo 'Pushing update...'
    push_files "$RELEASE_NAME"

    echo 'Creating pull request...'
    create_pr "$RELEASE_NAME"
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

push_files() {
    local release_name="$1"
    local branch_name="feature/${release_name}"
    local base_branch="gh-pages"

    # Fetch the latest state of the remote branches
    git fetch origin

    # Create a new branch from the latest commit of the gh-pages branch
    echo "Creating new branch '$branch_name' from '$base_branch'..."
    gh api -X POST /repos/solarwinds/swi-k8s-opentelemetry-collector/git/refs \
        --field ref="refs/heads/$branch_name" \
        --field sha="$(git rev-parse "origin/$base_branch")"

    # Get the SHA of the current index.yaml in the base branch
    SHA=$(gh api repos/solarwinds/swi-k8s-opentelemetry-collector/contents/index.yaml?ref="$base_branch" \
        --jq '.sha')

    MESSAGE="New release $release_name"

    # Push new index.yaml to the feature branch
    echo "Pushing new index.yaml to branch '$branch_name'..."
    gh api --method PUT /repos/solarwinds/swi-k8s-opentelemetry-collector/contents/index.yaml \
        --field message="$MESSAGE" \
        --field content=@<(base64 -i .cr-index/index.yaml) \
        --field branch="$branch_name" \
        --field sha="$SHA"
}

create_pr() {
    local release_name="$1"
    local branch_name="feature/${release_name}"
    local base_branch="gh-pages"

    # Create a pull request
    echo "Creating a pull request from '$branch_name' into '$base_branch'..."
    gh pr create --base "$base_branch" --head "$branch_name" \
        --title "Update Helm Chart for $release_name" \
        --body "This PR updates the Helm chart index.yaml with the latest release $release_name."
}

main "$@"