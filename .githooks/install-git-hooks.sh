#!/bin/bash

# Define the source and destination paths
source_path=".githooks/pre-push"
destination_path=".git/hooks/pre-push"

# Check if the source script exists
if [ -e "$source_path" ]; then
  # Copy the source script to the hooks directory
  cp "$source_path" "$destination_path"

  # Grant execution permission to the copied script
  chmod +x "$destination_path"

  echo "pre-push hook script has been successfully installed in .git/hooks/"
else
  echo "Error: The source pre-push script ('$source_path') does not exist. Please make sure it's in the correct location."
fi
