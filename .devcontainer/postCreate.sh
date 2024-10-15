#!/bin/sh

# Fix permissions for the Go cache
sudo chown -R $(id -u):$(id -g) /var/cache/go

# Install precommit (https://pre-commit.com/)
pre-commit install -t pre-commit -t prepare-commit-msg

# uv autocomplete
echo 'eval "$(uv generate-shell-completion zsh)"' >> ~/.zshrc

# uvx autocomplete
echo 'eval "$(uvx --generate-shell-completion zsh)"' >> ~/.zshrc
