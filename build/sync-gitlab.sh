#!/bin/bash
# GitLab Sync Script for Canvus Server DB Solver
# Usage: ./build/sync-gitlab.sh

set -e

echo "🔄 Syncing with GitLab repository..."

# Check if gitlab remote exists
if ! git remote get-url gitlab >/dev/null 2>&1; then
    echo "❌ GitLab remote not found. Adding it now..."
    git remote add gitlab https://gitlab.multitaction.com/swrd/canvus-server-db-solver.git
fi

# Fetch latest from GitLab
echo "📥 Fetching latest from GitLab..."
git fetch gitlab

# Push current branch to GitLab
echo "📤 Pushing to GitLab..."
git push gitlab master

# Also push tags if any exist
if git tag -l | grep -q .; then
    echo "🏷️  Pushing tags to GitLab..."
    git push gitlab --tags
fi

echo "✅ GitLab sync completed!"
echo ""
echo "💡 To sync from GitLab to local:"
echo "   git pull gitlab master"
echo ""
echo "💡 To check GitLab remote:"
echo "   git remote -v"
