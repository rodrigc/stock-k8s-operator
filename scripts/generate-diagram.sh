#!/bin/bash

# Check if mmdc is installed
if ! command -v mmdc &> /dev/null; then
    echo "Installing @mermaid-js/mermaid-cli..."
    npm install -g @mermaid-js/mermaid-cli
fi

# Create images directory if it doesn't exist
mkdir -p images

# Generate PNG from Mermaid file
mmdc -i images/architecture.mmd -o images/architecture.png -c mermaid-config.json 