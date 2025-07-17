#!/bin/bash
# Build script for Windows binary without external DLL dependencies

echo "Building Windows binary with CGO disabled..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o orochi_windows_amd64.exe -v ./cmd/orochi

echo "Windows binary built successfully: orochi_windows_amd64.exe"
echo "This binary should work without requiring libgcc_s_seh-1.dll"