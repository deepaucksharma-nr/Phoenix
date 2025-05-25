#!/bin/bash
# Phoenix CLI Installation Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="phoenix"
VERSION="latest"
ARCH=""
OS=""
DOWNLOAD_URL=""

# Print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$OS" in
        darwin)
            OS="darwin"
            ;;
        linux)
            OS="linux"
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    print_info "Detected platform: $OS/$ARCH"
}

# Check if running as root when installing to system directories
check_permissions() {
    if [[ "$INSTALL_DIR" == "/usr/local/bin" || "$INSTALL_DIR" == "/usr/bin" ]]; then
        if [[ $EUID -ne 0 ]]; then
            print_warning "Installing to $INSTALL_DIR requires sudo privileges"
            print_info "Re-running with sudo..."
            exec sudo "$0" "$@"
        fi
    fi
}

# Download the Phoenix CLI binary
download_cli() {
    local temp_dir=$(mktemp -d)
    local binary_file="$temp_dir/phoenix"
    
    print_info "Downloading Phoenix CLI..."
    
    # For now, we'll build from source since we don't have pre-built binaries
    # In production, this would download from GitHub releases
    if command -v go &> /dev/null; then
        print_info "Building Phoenix CLI from source..."
        
        # Save current directory
        local current_dir=$(pwd)
        
        # Change to phoenix-platform directory
        cd "$(dirname "$0")/.."
        
        # Build the CLI
        print_info "Running make build-cli..."
        make build-cli
        
        # Copy the built binary
        cp build/phoenix "$binary_file"
        
        # Return to original directory
        cd "$current_dir"
    else
        print_error "Go is not installed. Please install Go 1.21+ to build Phoenix CLI"
        print_info "Visit https://golang.org/dl/ for installation instructions"
        exit 1
    fi
    
    echo "$binary_file"
}

# Install the CLI binary
install_cli() {
    local binary_file=$1
    
    print_info "Installing Phoenix CLI to $INSTALL_DIR..."
    
    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"
    
    # Copy binary to install directory
    cp "$binary_file" "$INSTALL_DIR/$BINARY_NAME"
    
    # Make it executable
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    # Clean up temp file
    rm -f "$binary_file"
    
    print_info "Phoenix CLI installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        print_info "Verification successful!"
        print_info "Phoenix CLI version: $($BINARY_NAME version)"
    else
        print_error "Installation verification failed"
        print_info "Please ensure $INSTALL_DIR is in your PATH"
        print_info "You can add it by running:"
        print_info "  echo 'export PATH=\$PATH:$INSTALL_DIR' >> ~/.bashrc"
        print_info "  source ~/.bashrc"
    fi
}

# Setup shell completion
setup_completion() {
    print_info "Setting up shell completion..."
    
    # Detect shell
    if [[ -n "$BASH_VERSION" ]]; then
        local completion_dir=""
        if [[ -d "/etc/bash_completion.d" ]]; then
            completion_dir="/etc/bash_completion.d"
        elif [[ -d "/usr/local/etc/bash_completion.d" ]]; then
            completion_dir="/usr/local/etc/bash_completion.d"
        fi
        
        if [[ -n "$completion_dir" ]]; then
            print_info "Installing bash completion to $completion_dir"
            "$INSTALL_DIR/$BINARY_NAME" completion bash > "$completion_dir/phoenix"
            print_info "Bash completion installed. Restart your shell or run: source $completion_dir/phoenix"
        fi
    elif [[ -n "$ZSH_VERSION" ]]; then
        print_info "For zsh completion, add this to your ~/.zshrc:"
        print_info "  source <($BINARY_NAME completion zsh)"
    fi
}

# Main installation flow
main() {
    print_info "Phoenix CLI Installation Script"
    print_info "==============================="
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --install-dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --version)
                VERSION="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --install-dir DIR    Installation directory (default: /usr/local/bin)"
                echo "  --version VERSION    Version to install (default: latest)"
                echo "  --help              Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Detect platform
    detect_platform
    
    # Check permissions
    check_permissions
    
    # Download CLI
    binary_file=$(download_cli)
    
    # Install CLI
    install_cli "$binary_file"
    
    # Setup completion (optional)
    if [[ -t 0 ]]; then  # Only if running interactively
        read -p "Would you like to install shell completion? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            setup_completion
        fi
    fi
    
    # Verify installation
    verify_installation
    
    print_info ""
    print_info "Installation complete! 🎉"
    print_info ""
    print_info "Next steps:"
    print_info "  1. Run 'phoenix auth login' to authenticate"
    print_info "  2. Run 'phoenix --help' to see available commands"
    print_info "  3. Visit https://github.com/phoenix-platform/phoenix for documentation"
}

# Run main function
main "$@"