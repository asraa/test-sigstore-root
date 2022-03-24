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


# Ask user to insert key 
read -n1 -r -s -p "Insert your Yubikey, then press any key to continue..." 

# Sign the root and targets with hardware key
./tuf sign -repository $REPO -roles root -roles targets -sk

# Ask user to remove key (and replace with SSH security key)
read -n1 -r -s -p "Remove your Yubikey, then press any key to continue..." 

git checkout -b sign-root-targets
git add ceremony/
git commit -s -m "Signing root and targets for ${GITHUB_USER}"
git push -f origin sign-root-targets

# Open the browser
open "https://github.com/${GITHUB_USER}/test-sigstore-root/pull/new/sign-root-targets" || xdg-open "https://github.com/${GITHUB_USER}/test-sigstore-root/pull/new/sign-root-targets"
