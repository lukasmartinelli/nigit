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

# because we have the errexit option set we need to suppress exit codes
# from the linter as we still want to pass back the output
function suppress_lint_error() {
    local _="$?"
}

function lint_source() {
    local repo_path=$(clone_repo)

    if [ "$ACCEPT" == "application/json" ]; then
        hlint "$repo_path" --json || suppress_lint_error
    else
        hlint "$repo_path" --color=never || suppress_lint_error
    fi

    rm -rf "$repo_path"
}

lint_source
