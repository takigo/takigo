#!/bin/bash
# Run tests with race detector
# Usage: ./scripts/test_race.sh [packages...]

set -e

if [ $# -eq 0 ]; then
    go test -race ./...
else
    go test -race "$@"
fi