#!/bin/bash
set -Eeuo pipefail

# TODO: increase test scope
go test \
    `go list ./... | \
    grep "central/main\|central/lib\|central/spp" | \
    grep -v "central/main/handlers/assets" | \
    grep -v "central/main$" | \
    grep -v "central/spp/"`
