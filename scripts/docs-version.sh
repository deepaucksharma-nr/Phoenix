#!/bin/bash
# Documentation versioning script using mike

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() { echo -e "${GREEN}✓ $1${NC}"; }
print_error() { echo -e "${RED}✗ $1${NC}"; }
print_info() { echo -e "${YELLOW}→ $1${NC}"; }

# Check if mike is installed
if ! command -v mike &> /dev/null; then
    print_error "mike is not installed. Please install it with: pip install mike"
    exit 1
fi

# Parse command line arguments
COMMAND=${1:-}
VERSION=${2:-}

# Function to show usage
show_usage() {
    echo "Phoenix Documentation Versioning"
    echo ""
    echo "Usage: $0 <command> [version]"
    echo ""
    echo "Commands:"
    echo "  deploy <version>    Deploy a new version of the docs"
    echo "  list               List all deployed versions"
    echo "  set-default <ver>  Set the default version"
    echo "  delete <version>   Delete a version"
    echo "  retitle <ver> <title> Retitle a version"
    echo "  alias <ver> <alias>   Create an alias for a version"
    echo ""
    echo "Examples:"
    echo "  $0 deploy v1.0.0"
    echo "  $0 deploy v1.0.1 --push"
    echo "  $0 set-default v1.0.0"
    echo "  $0 alias v1.0.0 latest"
}

# Function to get current git tag
get_current_version() {
    # Try to get the current tag
    local tag=$(git describe --tags --exact-match 2>/dev/null || echo "")
    if [ -z "$tag" ]; then
        # If no tag, use branch name with commit hash
        local branch=$(git rev-parse --abbrev-ref HEAD)
        local commit=$(git rev-parse --short HEAD)
        echo "${branch}-${commit}"
    else
        echo "$tag"
    fi
}

# Function to deploy documentation
deploy_docs() {
    local version=$1
    shift # Remove version from arguments
    local extra_args="$@"
    
    if [ -z "$version" ]; then
        version=$(get_current_version)
        print_info "No version specified, using: $version"
    fi
    
    print_info "Building documentation..."
    mkdocs build --clean
    
    print_info "Deploying version $version..."
    
    # Deploy with mike
    if mike deploy "$version" $extra_args; then
        print_success "Successfully deployed version $version"
        
        # If this looks like a stable release, also tag it as latest
        if [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            print_info "Updating 'latest' alias to point to $version..."
            mike alias "$version" latest $extra_args
            print_success "Updated 'latest' alias"
        fi
    else
        print_error "Failed to deploy version $version"
        exit 1
    fi
}

# Function to list versions
list_versions() {
    print_info "Deployed documentation versions:"
    mike list
}

# Function to set default version
set_default() {
    local version=$1
    if [ -z "$version" ]; then
        print_error "Version is required"
        show_usage
        exit 1
    fi
    
    print_info "Setting default version to $version..."
    if mike set-default "$version" --push; then
        print_success "Successfully set default version to $version"
    else
        print_error "Failed to set default version"
        exit 1
    fi
}

# Function to delete a version
delete_version() {
    local version=$1
    if [ -z "$version" ]; then
        print_error "Version is required"
        show_usage
        exit 1
    fi
    
    print_info "Deleting version $version..."
    read -p "Are you sure you want to delete version $version? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if mike delete "$version" --push; then
            print_success "Successfully deleted version $version"
        else
            print_error "Failed to delete version"
            exit 1
        fi
    else
        print_info "Deletion cancelled"
    fi
}

# Function to retitle a version
retitle_version() {
    local version=$1
    local title=$2
    if [ -z "$version" ] || [ -z "$title" ]; then
        print_error "Both version and title are required"
        show_usage
        exit 1
    fi
    
    print_info "Retitling version $version to '$title'..."
    if mike retitle "$version" "$title" --push; then
        print_success "Successfully retitled version $version"
    else
        print_error "Failed to retitle version"
        exit 1
    fi
}

# Function to create an alias
create_alias() {
    local version=$1
    local alias=$2
    if [ -z "$version" ] || [ -z "$alias" ]; then
        print_error "Both version and alias are required"
        show_usage
        exit 1
    fi
    
    print_info "Creating alias '$alias' for version $version..."
    if mike alias "$version" "$alias" --push; then
        print_success "Successfully created alias '$alias'"
    else
        print_error "Failed to create alias"
        exit 1
    fi
}

# Main command handling
case "$COMMAND" in
    deploy)
        shift # Remove 'deploy' from arguments
        deploy_docs "$@"
        ;;
    list)
        list_versions
        ;;
    set-default)
        set_default "$VERSION"
        ;;
    delete)
        delete_version "$VERSION"
        ;;
    retitle)
        TITLE=${3:-}
        retitle_version "$VERSION" "$TITLE"
        ;;
    alias)
        ALIAS=${3:-}
        create_alias "$VERSION" "$ALIAS"
        ;;
    *)
        show_usage
        exit 1
        ;;
esac