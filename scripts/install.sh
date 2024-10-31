#!/bin/bash
# scripts/install.sh

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}Installing Nomad File Watcher Driver...${NC}"

# Create necessary directories
sudo mkdir -p /opt/nomad/plugins
sudo mkdir -p /var/lib/nomad/filewatcher

# Copy binary
sudo cp bin/nomad-filewatcher-driver /opt/nomad/plugins/

# Set permissions
sudo chmod 755 /opt/nomad/plugins/nomad-filewatcher-driver
sudo chown -R nomad:nomad /var/lib/nomad/filewatcher

# Create default configuration if it doesn't exist
if [ ! -f /etc/nomad.d/plugins/filewatcher.hcl ]; then
    sudo mkdir -p /etc/nomad.d/plugins
    sudo tee /etc/nomad.d/plugins/filewatcher.hcl > /dev/null <<EOF
plugin "filewatcher-driver" {
  config {
    enabled = true
    state_dir = "/var/lib/nomad/filewatcher"
    log_level = "INFO"
  }
}
EOF
fi

# Restart Nomad if it's running
if systemctl is-active --quiet nomad; then
    echo -e "${YELLOW}Restarting Nomad...${NC}"
    sudo systemctl restart nomad
fi

echo -e "${GREEN}Installation complete!${NC}"
echo -e "You can check the status with: nomad plugin status"