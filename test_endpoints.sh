#!/bin/bash
# test_endpoints.sh - Script to test SecretAI Attest REST Server endpoints
# This script uses curl to test all endpoints and formats the output

# Default server address
SERVER="https://localhost:29343"
# Default output format (json, raw, both)
FORMAT="both"
# Skip SSL verification by default (good for self-signed certs)
SKIP_SSL="true"

# Text colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to display usage
function show_usage {
    echo -e "${BLUE}Usage: $0 [options]${NC}"
    echo -e "  -s, --server URL  : Server URL (default: $SERVER)"
    echo -e "  -f, --format TYPE : Output format - json, raw, or both (default: $FORMAT)"
    echo -e "  -k, --ssl-verify  : Enable SSL verification (default: disabled)"
    echo -e "  -h, --help        : Show this help message"
    echo -e ""
    echo -e "Examples:"
    echo -e "  $0 --server https://myserver.example.com:8443"
    echo -e "  $0 --format json"
    echo -e "  $0 --ssl-verify"
    exit 1
}

# Parse command line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -s|--server) SERVER="$2"; shift ;;
        -f|--format) FORMAT="$2"; shift ;;
        -k|--ssl-verify) SKIP_SSL="false" ;;
        -h|--help) show_usage ;;
        *) echo "Unknown parameter: $1"; show_usage ;;
    esac
    shift
done

# Validate format
if [[ "$FORMAT" != "json" && "$FORMAT" != "raw" && "$FORMAT" != "both" ]]; then
    echo -e "${RED}Error: Format must be 'json', 'raw', or 'both'${NC}"
    show_usage
fi

# Set curl options
CURL_OPTS=""
if [[ "$SKIP_SSL" == "true" ]]; then
    CURL_OPTS="-k"
fi

# Function to print section headers
function print_header {
    echo -e "\n${GREEN}========== Testing $1 Endpoint ==========${NC}"
}

# Function to handle and display the curl response
function test_endpoint {
    local endpoint=$1
    local description=$2
    
    print_header "$description"
    
    echo -e "${YELLOW}Request: curl $CURL_OPTS $SERVER$endpoint${NC}"
    
    # Make the request
    local http_code=$(curl $CURL_OPTS -s -o response.tmp -w "%{http_code}" "$SERVER$endpoint")
    
    # Display HTTP code
    if [[ $http_code -ge 200 && $http_code -lt 300 ]]; then
        echo -e "${GREEN}Response Code: $http_code${NC}"
    else
        echo -e "${RED}Response Code: $http_code${NC}"
    fi
    
    # Display response based on format
    if [[ "$FORMAT" == "json" || "$FORMAT" == "both" ]]; then
        echo -e "${BLUE}Response (JSON):${NC}"
        if [[ -s response.tmp ]]; then
            # Check if it's JSON
            if python -c "import json; json.load(open('response.tmp'))" 2>/dev/null; then
                python -m json.tool response.tmp | head -20
                lines=$(wc -l < response.tmp)
                if [[ $lines -gt 20 ]]; then
                    echo -e "${YELLOW}... (response truncated, showing first 20 lines)${NC}"
                fi
            else
                echo -e "${YELLOW}Response is not valid JSON. Showing raw output:${NC}"
                head -20 response.tmp
                lines=$(wc -l < response.tmp)
                if [[ $lines -gt 20 ]]; then
                    echo -e "${YELLOW}... (response truncated, showing first 20 lines)${NC}"
                fi
            fi
        else
            echo -e "${RED}Empty response${NC}"
        fi
    fi
    
    if [[ "$FORMAT" == "raw" || "$FORMAT" == "both" ]]; then
        echo -e "${BLUE}Response (Raw - first 20 lines):${NC}"
        if [[ -s response.tmp ]]; then
            head -20 response.tmp
            lines=$(wc -l < response.tmp)
            if [[ $lines -gt 20 ]]; then
                echo -e "${YELLOW}... (response truncated, showing first 20 lines)${NC}"
            fi
        else
            echo -e "${RED}Empty response${NC}"
        fi
    fi
    
    # Clean up
    rm -f response.tmp
}

# Test each endpoint
echo -e "${BLUE}SecretAI Attest REST Server Endpoint Test${NC}"
echo -e "${YELLOW}Server: $SERVER${NC}"
echo -e "${YELLOW}Format: $FORMAT${NC}"
echo -e "${YELLOW}SSL Verification: $(if [[ "$SKIP_SSL" == "true" ]]; then echo "Disabled"; else echo "Enabled"; fi)${NC}"

# Test /status
test_endpoint "/status" "Status"

# Test /attestation
test_endpoint "/attestation" "Attestation"

# Test /gpu
test_endpoint "/gpu" "GPU Attestation"

# Test /cpu
test_endpoint "/cpu" "CPU Attestation"

# Test /self
test_endpoint "/self" "Self Attestation"

echo -e "\n${GREEN}========== Testing Complete ==========${NC}"
echo -e "${BLUE}For more details, check the server logs${NC}"