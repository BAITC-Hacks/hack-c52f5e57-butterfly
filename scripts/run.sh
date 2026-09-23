#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
mkdir -p bin
go build -o bin/butterfly ./cmd/butterfly
exec ./bin/butterfly
