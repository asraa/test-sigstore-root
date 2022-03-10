#!/bin/bash

# Print all commands and stop on errors
set -ex

if [ -z "$GITHUB_USER" ]; then
    echo "Set GITHUB_USER"
    exit
fi
if [ -z "$REPO" ]; then
    REPO=$(pwd)/ceremony/$(date '%Y-%m-%d')
    echo "Using default REPO $REPO"
fi

# Dump the git state
git checkout main
git status
git remote -v

# Setup forks
git remote rm upstream || true
git remote add upstream git@github.com:asraa/test-sigstore-root.git
git remote rm origin || true
git remote add origin git@github.com:"$GITHUB_USER"/test-sigstore-root.git
git remote -v

# build the verification binary
go build -o verify ./cmd/verify
[ -f piv-attestation-ca.pem ] || curl -fsO https://developers.yubico.com/PIV/Introduction/piv-attestation-ca.pem

# Fetch the pull request if specified and verify
if [[ ! -z "$1" ]]; then
    # Pull request to verify. If not supplied, use main
    echo "Pull Request: $1"
    git branch -D VERIFY || true
    git fetch upstream pull/$1/head:VERIFY
    git checkout VERIFY
fi

./verify keys --root piv-attestation-ca.pem --key-directory $REPO/keys
./verify repository --repository $REPO --root $REPO/repository/1.root.json

# stay on the branch for manual verification

