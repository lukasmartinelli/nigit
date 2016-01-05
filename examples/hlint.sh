#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

readonly ACCEPT=${ACCEPT:-text/plain}

function clone_repo() {
    local working_dir=$(mktemp -dt "hlint")
    local git_output=$(git clone --quiet "$GIT_REPOSITORY" "$working_dir")
    echo "$git_output" >&2
    echo "$working_dir"
}

function lint_source() {
    local repo_path=$(clone_repo)

    if [ "$ACCEPT" == "application/json" ]; then
        hlint "$repo_path" --json
    else
        hlint "$repo_path" --color=never
    fi

    rm -rf "$repo_path"
}

lint_source
