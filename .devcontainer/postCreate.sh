#!/bin/sh

# Fix permissions for the Go cache
sudo chown -R $(id -u):$(id -g) /var/cache/go

# Install precommit (https://pre-commit.com/)
pre-commit install -t pre-commit -t prepare-commit-msg

mkdir -p ~/.oh-my-zsh/custom/completions

# uv autocomplete
uv generate-shell-completion zsh >> ~/.oh-my-zsh/custom/completions/_uv

# uvx autocomplete
uvx --generate-shell-completion zsh >> ~/.oh-my-zsh/custom/completions/_uvx
