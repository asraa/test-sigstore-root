#!/bin/bash

# Print all commands and stop on errors
set -ex

if [ -z "$GITHUB_USER" ]; then
    echo "Set GITHUB_USER"
    exit
fi
if [ -z "$REPO" ]; then
    REPO=$(pwd)/ceremony/$(date '+%Y-%m-%d')
    echo "Using default REPO $REPO"
fi

# Dump the git state
git status
git remote -v

git clean -d -f
git checkout main
git pull upstream main
git status

# Ask user to insert key 
read -n1 -r -s -p "Insert your Yubikey, then press any key to continue...\n" 

# Add the key!
./tuf add-key -repository $REPO

# Ask user to remove key (and replace with SSH security key)
read -n1 -r -s -p "Remove your Yubikey, then press any key to continue...\n" 

git status
git checkout -b add-key
git add ceremony/
git commit -s -a -m "Adding initial key for ${GITHUB_USER}"
git push -f origin add-key

# Open the browser
export GITHUB_URL=$(git remote -v | awk '/^upstream/{print $2}'| head -1 | sed -Ee 's#(git@|git://)#https://#' -e 's@com:@com/@' -e 's#\.git$##')
export BRANCH=$(git symbolic-ref HEAD | cut -d"/" -f 3,4)
export PR_URL=${GITHUB_URL}"/compare/main..."${BRANCH}"?expand=1"
open "${PR_URL}" || xdg-open "${PR_URL}"
