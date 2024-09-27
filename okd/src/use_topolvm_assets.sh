#!/bin/bash
set -x

REPO_DIR=$(readlink -nf "$(dirname "${BASE_SOURCE}")/../../")
ASSET_DIR="${REPO_DIR}/okd/assets/components/storage"
RELEASE_DATA="${ASSET_DIR}/release-x86_64.json"
OUT_DIR="${REPO_DIR}/_output/topolvm"

mkdir -p "${OUT_DIR}"

gomplate \
    --input-dir="${ASSET_DIR}" \
    --context="ReleaseImage=${RELEASE_DATA}" \
    --output-dir="${OUT_DIR}"