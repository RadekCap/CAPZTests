#!/usr/bin/env bash
# cleanup-azure-resources.sh - Clean up Azure resources created during ARO-CAPZ testing
#
# This script finds and deletes Azure resources that match the naming patterns used
# during testing. These resources may not be tied to the resource group and can
# survive resource group deletion.
#
# Usage:
#   ./scripts/cleanup-azure-resources.sh [OPTIONS]
#
# Options:
#   --prefix PREFIX    Resource name prefix to search for (default: from CAPZ_USER env var or 'rcap')
#   --dry-run          Show what would be deleted without actually deleting
#   --force            Skip confirmation prompts
#   --help             Show this help message
#
# Environment variables:
#   CAPZ_USER          Default prefix for resource names (e.g., 'rcap')
#   AZURE_SUBSCRIPTION_ID  Azure subscription ID to search in
#
# Examples:
#   ./scripts/cleanup-azure-resources.sh --dry-run
#   ./scripts/cleanup-azure-resources.sh --prefix rcapd --force
#   CAPZ_USER=myuser ./scripts/cleanup-azure-resources.sh

set -euo pipefail

# Default values
PREFIX="${CAPZ_USER:-rcap}"
DRY_RUN=false
FORCE=false

# ANSI colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[OK]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Show usage
usage() {
    head -30 "$0" | grep '^#' | sed 's/^# \?//'
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --prefix)
            PREFIX="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        --help|-h)
            usage
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."

    if ! command -v az >/dev/null 2>&1; then
        print_error "Azure CLI (az) is not installed"
        echo "Install from: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
        exit 1
    fi

    if ! az account show >/dev/null 2>&1; then
        print_error "Not logged in to Azure CLI"
        echo "Run 'az login' to authenticate"
        exit 1
    fi

    # Check for resource-graph extension
    if ! az extension show --name resource-graph >/dev/null 2>&1; then
        print_warning "Azure Resource Graph extension not installed"
        print_info "Installing resource-graph extension..."
        az extension add --name resource-graph --yes
    fi

    print_success "Prerequisites check passed"
}

# Find resources matching the pattern
find_resources() {
    local prefix="$1"

    print_info "Searching for Azure resources with prefix '${prefix}'..."

    # Query Azure Resource Graph for resources matching the pattern
    # We search for:
    # 1. Resources with names starting with the prefix (e.g., rcapa, rcapb, rcapc, etc.)
    # 2. Resources with names containing the prefix pattern

    # Build the query to find resources with the naming pattern
    # The pattern is: prefix followed by optional suffix (e.g., rcapa, rcapb, rcap-stage, etc.)
    local query="Resources | where name contains '${prefix}' | project id, name, type, resourceGroup, subscriptionId | order by type asc, name asc"

    az graph query -q "$query" -o json 2>/dev/null
}

# Parse and display resources
display_resources() {
    local resources_json="$1"
    local count

    # Handle empty or invalid JSON
    if [[ -z "$resources_json" ]] || ! echo "$resources_json" | jq -e '.' >/dev/null 2>&1; then
        print_info "No resources found matching prefix '${PREFIX}'"
        return 1
    fi

    count=$(echo "$resources_json" | jq -r '.data | length // 0')

    if [[ "$count" -eq 0 ]]; then
        print_info "No resources found matching prefix '${PREFIX}'"
        return 1
    fi

    echo ""
    print_warning "Found ${count} resource(s) matching prefix '${PREFIX}':"
    echo ""

    # Print table header
    printf "%-60s | %-50s | %-30s\n" "NAME" "TYPE" "RESOURCE GROUP"
    printf "%s\n" "$(printf '%.0s-' {1..145})"

    # Print each resource
    echo "$resources_json" | jq -r '.data[] | "\(.name)|\(.type)|\(.resourceGroup)"' | while IFS='|' read -r name type rg; do
        # Truncate long names for display
        name_display="${name:0:60}"
        type_display="${type:0:50}"
        rg_display="${rg:0:30}"
        printf "%-60s | %-50s | %-30s\n" "$name_display" "$type_display" "$rg_display"
    done

    echo ""
    return 0
}

# Delete resources
delete_resources() {
    local resources_json="$1"
    local count
    local deleted=0
    local failed=0
    local skipped=0

    count=$(echo "$resources_json" | jq -r '.data | length')

    if [[ "$count" -eq 0 ]]; then
        return 0
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        print_warning "[DRY-RUN] Would delete ${count} resource(s)"
        return 0
    fi

    # Confirm deletion unless --force is specified
    if [[ "$FORCE" != "true" ]]; then
        echo ""
        read -p "Delete all ${count} resource(s)? [y/N] " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Deletion cancelled"
            return 0
        fi
    fi

    echo ""
    print_info "Deleting resources..."
    echo ""

    # Get resource IDs and delete them
    # Sort by type to handle dependencies (identities first, then VNets, then NSGs)
    local resource_ids
    resource_ids=$(echo "$resources_json" | jq -r '.data | sort_by(.type) | reverse | .[].id')

    while IFS= read -r resource_id; do
        if [[ -z "$resource_id" ]]; then
            continue
        fi

        local resource_name
        resource_name=$(basename "$resource_id")

        echo -n "  Deleting: ${resource_name}... "

        # First verify the resource still exists (Resource Graph may have stale data)
        if ! az resource show --ids "$resource_id" >/dev/null 2>&1; then
            echo "SKIPPED (not found)"
            ((skipped++))
            continue
        fi

        # Attempt deletion
        if az resource delete --ids "$resource_id" --no-wait 2>/dev/null; then
            echo "INITIATED"
            ((deleted++))
        else
            echo "FAILED"
            ((failed++))
        fi
    done <<< "$resource_ids"

    echo ""
    print_info "Deletion summary:"
    echo "  - Initiated: ${deleted}"
    echo "  - Failed: ${failed}"
    echo "  - Skipped (not found): ${skipped}"

    if [[ "$deleted" -gt 0 ]]; then
        print_warning "Note: Deletions run asynchronously. Resources may take a few minutes to be fully removed."
        print_info "Run this script again to verify cleanup is complete."
    fi
}

# Main function
main() {
    echo "========================================"
    echo "=== Azure Resource Cleanup ==="
    echo "========================================"
    echo ""
    print_info "Resource prefix: ${PREFIX}"
    if [[ "$DRY_RUN" == "true" ]]; then
        print_warning "DRY-RUN mode enabled - no resources will be deleted"
    fi
    echo ""

    check_prerequisites
    echo ""

    # Find resources
    local resources_json
    resources_json=$(find_resources "$PREFIX")

    # Display found resources
    if ! display_resources "$resources_json"; then
        print_success "No cleanup needed"
        exit 0
    fi

    # Delete resources
    delete_resources "$resources_json"

    echo ""
    echo "========================================"
    echo "=== Cleanup Complete ==="
    echo "========================================"
}

main "$@"
