#!/bin/bash
# Medha MCP Server Wrapper with Environment Variables
# This ensures the database path is always correctly set

# Set database path explicitly using user's home directory
export DB_PATH="$HOME/.medha/db/medha.db"

# Optional: Set encryption key if needed
# export ENCRYPTION_KEY="your-key-here"

# Run medha with the environment variables set
exec "$(dirname "$0")/bin/medha" "$@"
