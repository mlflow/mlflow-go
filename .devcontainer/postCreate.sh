#!/bin/sh

# Install precommit (https://pre-commit.com/)
pre-commit install -t pre-commit -t prepare-commit-msg

# Install Mage (https://magefile.org/)
go install github.com/magefile/mage@v1.15.0
