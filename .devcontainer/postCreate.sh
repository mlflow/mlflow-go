!#/bin/sh

# Install precommit (https://pre-commit.com/)
pre-commit install -t pre-commit -t prepare-commit-msg

# Download Go dependencies
go mod tidy

# Install Mage (https://magefile.org/)
go install github.com/magefile/mage@v1.15.0
