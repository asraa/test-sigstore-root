#!/bin/bash

# Print all commands and stop on errors
set -ex

if [ -z "$GITHUB_USER" ]; then
    echo "Set GITHUB_USER"
    exit
fi
if [ -z "$CEREMONY_DATE" ]; then
    CEREMONY_DATE=$(date '+%Y-%m-%d')
fi
export REPO=$(pwd)/ceremony/$CEREMONY_DATE

# Dump the git state
git status
git remote -v

git clean -d -f
git checkout main
git pull upstream main
git status

# Add the key!
./tuf add-key -repository $REPO
git status
git checkout -b add-key
git add ceremony/
git commit -s -a -m "Adding initial key for ${GITHUB_USER}"
git push -f origin add-key

# Open the browser
open "https://github.com/${GITHUB_USER}/test-sigstore-root/pull/new/add-key" || xdg-open "https://github.com/${GITHUB_USER}/test-sigstore-root/pull/new/add-key"
