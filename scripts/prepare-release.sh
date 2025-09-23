#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if version argument is provided
if [ $# -eq 0 ]; then
    print_error "Usage: $0 <version>"
    print_error "Example: $0 v1.0.0"
    exit 1
fi

VERSION=$1

# Validate semantic version format (vX.Y.Z)
if ! [[ $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    print_error "Version must be in format vX.Y.Z (e.g., v1.0.0)"
    exit 1
fi

# Extract version without 'v' prefix for version.go
VERSION_NUMBER=${VERSION#v}

print_status "Preparing release $VERSION"

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    print_error "Working directory is not clean. Please commit or stash changes."
    exit 1
fi

# Get current branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

# Warn if on main branch
if [ "$CURRENT_BRANCH" == "main" ]; then
    print_warning "You're on the main branch!"
    print_warning "This script should be run on a feature branch."
    print_warning "The GitHub Actions workflow will auto-tag when merged to main."
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Release preparation cancelled"
        exit 0
    fi
fi

# Get current version from version.go
CURRENT_VERSION=$(grep 'const Version = ' internal/version.go | sed 's/const Version = "\(.*\)"/\1/')
print_status "Current version: $CURRENT_VERSION"
print_status "New version: $VERSION_NUMBER"

# Update version.go
print_status "Updating internal/version.go"
sed -i.bak "s/const Version = \"[^\"]*\"/const Version = \"$VERSION_NUMBER\"/" internal/version.go

# Remove backup file
rm internal/version.go.bak

# Show the change
print_status "Version file updated:"
grep "const Version" internal/version.go

# Commit the version bump
print_status "Committing version bump"
git add internal/version.go
git commit -m "Bump version to $VERSION

Prepare for release $VERSION

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"

print_status "✅ Version bump committed on branch: $CURRENT_BRANCH"
print_status ""
print_status "Next steps:"
print_status "1. Push this branch: git push origin $CURRENT_BRANCH"
print_status "2. Create PR to merge into main"
print_status "3. After PR is merged, GitHub Actions will automatically:"
print_status "   - Detect version change"
print_status "   - Create git tag: $VERSION"
print_status "   - Create GitHub release"
print_status "   - Make available: go get github.com/ReforgeHQ/sdk-go@$VERSION"

print_status "Done!"