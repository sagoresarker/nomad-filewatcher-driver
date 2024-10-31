#!/bin/bash
# scripts/build.sh

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Get the directory of this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "${DIR}/.."

echo -e "${YELLOW}Building Nomad File Watcher Driver...${NC}"

# Create build directory
mkdir -p bin

# Run tests
echo -e "${YELLOW}Running tests...${NC}"
go test ./... || {
    echo -e "${RED}Tests failed${NC}"
    exit 1
}

# Build the binary
echo -e "${YELLOW}Building binary...${NC}"
go build -o bin/nomad-filewatcher-driver ./cmd/nomad-filewatcher-driver

# Create plugins directory if it doesn't exist
sudo mkdir -p /opt/nomad/plugins

echo -e "${GREEN}Build complete! Binary is in bin/nomad-filewatcher-driver${NC}"