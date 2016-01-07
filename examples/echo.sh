#!/bin/bash
set -o errexit
set -o pipefail
set -o nounset

input=$(cat)
echo "$input"
