#!/bin/bash
# secrets-manager.sh - Tool for secure secrets management in Phoenix Platform
# Created by Abhinav as part of Security & Compliance task

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_DIR="$REPO_ROOT/configs"
SECRETS_DIR="$CONFIG_DIR/secrets"

# Detect cloud environment
if [[ -n "${AWS_EXECUTION_ENV:-}" ]]; then
  CLOUD_ENV="aws"
elif [[ -n "${GOOGLE_CLOUD_PROJECT:-}" ]]; then
  CLOUD_ENV="gcp"
elif [[ -n "${AZURE_HTTP_USER_AGENT:-}" ]]; then
  CLOUD_ENV="azure"
else
  CLOUD_ENV="local"
fi

# Ensure required tools are installed
check_dependencies() {
  echo -e "${BLUE}Checking dependencies...${NC}"
  
  local missing_deps=false
  
  if ! command -v jq &> /dev/null; then
    echo -e "${RED}jq is required but not installed${NC}"
    missing_deps=true
  fi
  
  if ! command -v gpg &> /dev/null; then
    echo -e "${RED}gpg is required but not installed${NC}"
    missing_deps=true
  fi
  
  if ! command -v base64 &> /dev/null; then
    echo -e "${RED}base64 is required but not installed${NC}"
    missing_deps=true
  fi
  
  if [[ "$missing_deps" == "true" ]]; then
    echo -e "${RED}Missing dependencies. Install them and try again.${NC}"
    exit 1
  fi
  
  echo -e "${GREEN}All dependencies satisfied${NC}"
}

# Initialize the secrets directory structure
init() {
  echo -e "${BLUE}Initializing secrets management...${NC}"
  
  mkdir -p "$SECRETS_DIR"
  mkdir -p "$SECRETS_DIR/dev"
  mkdir -p "$SECRETS_DIR/staging"
  mkdir -p "$SECRETS_DIR/prod"
  
  # Create a .gitignore file to ensure secret files aren't committed
  if [ ! -f "$SECRETS_DIR/.gitignore" ]; then
    echo "*" > "$SECRETS_DIR/.gitignore"
    echo "!.gitignore" >> "$SECRETS_DIR/.gitignore"
    echo "!README.md" >> "$SECRETS_DIR/.gitignore"
  fi
  
  # Create a README file explaining usage
  if [ ! -f "$SECRETS_DIR/README.md" ]; then
    cat > "$SECRETS_DIR/README.md" << EOF
# Phoenix Platform Secrets Management

This directory contains encrypted secrets for the Phoenix Platform. The directory structure is:

\`\`\`
secrets/
├── dev/           # Development environment secrets
├── staging/       # Staging environment secrets
└── prod/          # Production environment secrets
\`\`\`

## Using the Secrets Manager

To manage secrets, use the \`secrets-manager.sh\` script:

\`\`\`bash
# Set a secret
./scripts/secrets-manager.sh set dev database.password "your-secret-value"

# Get a secret
./scripts/secrets-manager.sh get dev database.password

# List all secrets in an environment
./scripts/secrets-manager.sh list dev

# Rotate a secret
./scripts/secrets-manager.sh rotate dev database.password

# Deploy secrets to the environment
./scripts/secrets-manager.sh deploy dev
\`\`\`

## Security Notes

1. Never commit unencrypted secrets to version control
2. Treat encryption keys and access credentials with care
3. Rotate secrets regularly (at least quarterly)
4. Limit who has access to production secrets
\`\`\`
EOF
  fi
  
  echo -e "${GREEN}Secrets management initialized at $SECRETS_DIR${NC}"
}

# Encrypt a value with a key
encrypt() {
  local env=$1
  local key=$2
  local value=$3
  
  if [[ "$CLOUD_ENV" == "local" ]]; then
    # Local development encryption using GPG
    echo -n "$value" | gpg --symmetric --batch --passphrase "${PHOENIX_ENCRYPTION_KEY:-phoenix}" --quiet -o "$SECRETS_DIR/$env/$key.gpg"
  else
    # Cloud-based encryption
    case "$CLOUD_ENV" in
      aws)
        # AWS encryption using KMS
        aws_kms_key_id="${AWS_KMS_KEY_ID:-alias/phoenix-$env}"
        echo -n "$value" | aws kms encrypt --key-id "$aws_kms_key_id" --plaintext fileb:///dev/stdin --output text --query CiphertextBlob > "$SECRETS_DIR/$env/$key.enc"
        ;;
      gcp)
        # GCP encryption using KMS
        gcp_kms_key="${GCP_KMS_KEY:-projects/$GOOGLE_CLOUD_PROJECT/locations/global/keyRings/phoenix/cryptoKeys/phoenix-$env}"
        echo -n "$value" | gcloud kms encrypt --key "$gcp_kms_key" --plaintext-file - --ciphertext-file "$SECRETS_DIR/$env/$key.enc"
        ;;
      azure)
        # Azure encryption using Key Vault
        azure_keyvault="${AZURE_KEYVAULT:-phoenix-kv}"
        azure_key="${AZURE_KEY:-phoenix-$env}"
        echo -n "$value" | az keyvault key encrypt --vault-name "$azure_keyvault" --name "$azure_key" --algorithm RSA-OAEP-256 --value "$(cat -)" --output tsv > "$SECRETS_DIR/$env/$key.enc"
        ;;
    esac
  fi
}

# Decrypt a value with a key
decrypt() {
  local env=$1
  local key=$2
  
  if [ ! -f "$SECRETS_DIR/$env/$key.gpg" ] && [ ! -f "$SECRETS_DIR/$env/$key.enc" ]; then
    echo -e "${RED}Secret $key not found in $env environment${NC}"
    return 1
  fi
  
  if [[ "$CLOUD_ENV" == "local" ]]; then
    # Local development decryption using GPG
    gpg --batch --quiet --decrypt --passphrase "${PHOENIX_ENCRYPTION_KEY:-phoenix}" "$SECRETS_DIR/$env/$key.gpg" 2>/dev/null
  else
    # Cloud-based decryption
    case "$CLOUD_ENV" in
      aws)
        # AWS decryption using KMS
        aws kms decrypt --ciphertext-blob "fileb://$SECRETS_DIR/$env/$key.enc" --output text --query Plaintext | base64 --decode
        ;;
      gcp)
        # GCP decryption using KMS
        gcloud kms decrypt --key "${GCP_KMS_KEY:-projects/$GOOGLE_CLOUD_PROJECT/locations/global/keyRings/phoenix/cryptoKeys/phoenix-$env}" --ciphertext-file "$SECRETS_DIR/$env/$key.enc" --plaintext-file -
        ;;
      azure)
        # Azure decryption using Key Vault
        az keyvault key decrypt --vault-name "${AZURE_KEYVAULT:-phoenix-kv}" --name "${AZURE_KEY:-phoenix-$env}" --algorithm RSA-OAEP-256 --value "$(cat "$SECRETS_DIR/$env/$key.enc")" --output tsv | base64 --decode
        ;;
    esac
  fi
}

# Set a secret
set_secret() {
  local env=$1
  local key=$2
  local value=$3
  
  if [ ! -d "$SECRETS_DIR/$env" ]; then
    echo -e "${RED}Invalid environment: $env${NC}"
    exit 1
  fi
  
  echo -e "${BLUE}Setting secret $key in $env environment...${NC}"
  encrypt "$env" "$key" "$value"
  echo -e "${GREEN}Secret set successfully${NC}"
}

# Get a secret
get_secret() {
  local env=$1
  local key=$2
  
  if [ ! -d "$SECRETS_DIR/$env" ]; then
    echo -e "${RED}Invalid environment: $env${NC}"
    exit 1
  fi
  
  echo -e "${BLUE}Getting secret $key from $env environment...${NC}"
  decrypt "$env" "$key"
}

# List all secrets in an environment
list_secrets() {
  local env=$1
  
  if [ ! -d "$SECRETS_DIR/$env" ]; then
    echo -e "${RED}Invalid environment: $env${NC}"
    exit 1
  fi
  
  echo -e "${BLUE}Secrets in $env environment:${NC}"
  find "$SECRETS_DIR/$env" -type f -name "*.gpg" -o -name "*.enc" | sed "s#$SECRETS_DIR/$env/##" | sed 's/\.[^.]*$//' | sort
}

# Rotate a secret
rotate_secret() {
  local env=$1
  local key=$2
  local new_value=$3
  
  if [ ! -d "$SECRETS_DIR/$env" ]; then
    echo -e "${RED}Invalid environment: $env${NC}"
    exit 1
  fi
  
  if [ ! -f "$SECRETS_DIR/$env/$key.gpg" ] && [ ! -f "$SECRETS_DIR/$env/$key.enc" ]; then
    echo -e "${RED}Secret $key not found in $env environment${NC}"
    exit 1
  fi
  
  echo -e "${BLUE}Rotating secret $key in $env environment...${NC}"
  
  # Backup the current secret
  local timestamp=$(date +%Y%m%d%H%M%S)
  if [ -f "$SECRETS_DIR/$env/$key.gpg" ]; then
    cp "$SECRETS_DIR/$env/$key.gpg" "$SECRETS_DIR/$env/$key.$timestamp.gpg.bak"
  else
    cp "$SECRETS_DIR/$env/$key.enc" "$SECRETS_DIR/$env/$key.$timestamp.enc.bak"
  fi
  
  # Set the new secret value
  if [ -z "$new_value" ]; then
    # Generate a random secure password if no value provided
    new_value=$(openssl rand -base64 32)
    echo -e "${YELLOW}Generated random secret for $key${NC}"
  fi
  
  encrypt "$env" "$key" "$new_value"
  echo -e "${GREEN}Secret rotated successfully${NC}"
  
  # Return the new value
  echo "$new_value"
}

# Deploy secrets to Kubernetes
deploy_secrets() {
  local env=$1
  
  if [ ! -d "$SECRETS_DIR/$env" ]; then
    echo -e "${RED}Invalid environment: $env${NC}"
    exit 1
  fi
  
  echo -e "${BLUE}Deploying secrets to $env environment...${NC}"
  
  # Get the current Kubernetes context
  local current_context=$(kubectl config current-context)
  
  # Check if the context matches the environment
  if [[ ! "$current_context" =~ $env ]]; then
    echo -e "${YELLOW}Warning: Current Kubernetes context ($current_context) may not match target environment ($env)${NC}"
    read -p "Continue? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo -e "${RED}Deployment aborted${NC}"
      exit 1
    fi
  fi
  
  # Create a temporary file for Kubernetes secrets
  local temp_file=$(mktemp)
  trap "rm -f $temp_file" EXIT
  
  # Start the Kubernetes secret YAML
  cat > "$temp_file" << EOF
apiVersion: v1
kind: Secret
metadata:
  name: phoenix-secrets
  namespace: phoenix-$env
type: Opaque
data:
EOF
  
  # Add each secret to the YAML
  for secret_file in $(find "$SECRETS_DIR/$env" -type f -name "*.gpg" -o -name "*.enc" | sort); do
    local key=$(basename "$secret_file" | sed 's/\.[^.]*$//')
    local value=$(get_secret "$env" "$key" | base64 -w 0)
    echo "  $key: $value" >> "$temp_file"
  done
  
  # Apply the secret to Kubernetes
  kubectl apply -f "$temp_file"
  
  echo -e "${GREEN}Secrets deployed successfully${NC}"
}

# Display help
show_help() {
  cat << EOF
Phoenix Platform Secrets Manager

Usage: $0 <command> [arguments]

Commands:
  init                         Initialize secrets directory structure
  set <env> <key> <value>      Set a secret
  get <env> <key>              Get a secret
  list <env>                   List all secrets in an environment
  rotate <env> <key> [value]   Rotate a secret (generates random if value not provided)
  deploy <env>                 Deploy secrets to the environment

Environments:
  dev, staging, prod

Examples:
  $0 init
  $0 set dev database.password "secure-password"
  $0 get dev database.password
  $0 list dev
  $0 rotate dev database.password
  $0 deploy dev

For cloud environments, ensure appropriate credentials are configured.
EOF
}

# Main function
main() {
  check_dependencies
  
  if [ $# -lt 1 ]; then
    show_help
    exit 1
  fi
  
  command=$1
  shift
  
  case "$command" in
    init)
      init
      ;;
    set)
      if [ $# -lt 3 ]; then
        echo -e "${RED}Missing arguments. Usage: $0 set <env> <key> <value>${NC}"
        exit 1
      fi
      set_secret "$1" "$2" "$3"
      ;;
    get)
      if [ $# -lt 2 ]; then
        echo -e "${RED}Missing arguments. Usage: $0 get <env> <key>${NC}"
        exit 1
      fi
      get_secret "$1" "$2"
      ;;
    list)
      if [ $# -lt 1 ]; then
        echo -e "${RED}Missing environment. Usage: $0 list <env>${NC}"
        exit 1
      fi
      list_secrets "$1"
      ;;
    rotate)
      if [ $# -lt 2 ]; then
        echo -e "${RED}Missing arguments. Usage: $0 rotate <env> <key> [value]${NC}"
        exit 1
      fi
      rotate_secret "$1" "$2" "${3:-}"
      ;;
    deploy)
      if [ $# -lt 1 ]; then
        echo -e "${RED}Missing environment. Usage: $0 deploy <env>${NC}"
        exit 1
      fi
      deploy_secrets "$1"
      ;;
    help)
      show_help
      ;;
    *)
      echo -e "${RED}Unknown command: $command${NC}"
      show_help
      exit 1
      ;;
  esac
}

main "$@"
