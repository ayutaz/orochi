#!/bin/bash
# Script to test binary execution and check for DLL dependencies

set -e

echo "=== Binary Runtime Test Script ==="
echo "This script tests if binaries can run without external dependencies"
echo ""

# Function to test binary execution
test_binary() {
    local binary=$1
    local platform=$2
    
    echo "Testing $platform binary: $binary"
    
    if [ ! -f "$binary" ]; then
        echo "  ❌ Binary not found"
        return 1
    fi
    
    # Make executable
    chmod +x "$binary"
    
    # Test basic execution
    if "$binary" --help > /dev/null 2>&1; then
        echo "  ✅ Binary runs successfully"
        return 0
    else
        echo "  ❌ Binary failed to execute"
        echo "  This might be due to missing DLL dependencies on Windows"
        return 1
    fi
}

# Build Windows binary with CGO disabled
echo "Building Windows binary with CGO disabled..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o orochi_test_windows.exe ./cmd/orochi

# Build Linux binary
echo "Building Linux binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o orochi_test_linux ./cmd/orochi

# Build macOS binaries
echo "Building macOS binaries..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o orochi_test_darwin_amd64 ./cmd/orochi
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o orochi_test_darwin_arm64 ./cmd/orochi

echo ""
echo "=== Binary Analysis ==="

# Check file sizes
echo "Binary sizes:"
ls -lh orochi_test_* | awk '{print "  " $9 ": " $5}'

echo ""
echo "=== Checking Windows binary for DLL dependencies ==="

# Use 'file' command if available to check binary type
if command -v file &> /dev/null; then
    echo "File type analysis:"
    file orochi_test_windows.exe | sed 's/^/  /'
fi

# Check if strings contain references to problematic DLLs
if command -v strings &> /dev/null; then
    echo ""
    echo "Checking for DLL references in Windows binary:"
    problematic_dlls=("libgcc" "libwinpthread" "libstdc++" "msvcr" "cygwin")
    found_dll=false
    
    for dll in "${problematic_dlls[@]}"; do
        if strings orochi_test_windows.exe | grep -i "$dll" > /dev/null; then
            echo "  ⚠️  Found reference to $dll"
            found_dll=true
        fi
    done
    
    if [ "$found_dll" = false ]; then
        echo "  ✅ No references to problematic DLLs found"
    fi
fi

echo ""
echo "=== Summary ==="
echo "Binaries built with CGO_ENABLED=0 should work without external DLLs."
echo "The Windows binary (orochi_test_windows.exe) can now be tested on a Windows system."
echo ""
echo "To add this check to CI, create a GitHub workflow that:"
echo "1. Builds binaries with CGO_ENABLED=0"
echo "2. Tests execution with --help flag"
echo "3. On Windows, runs in a clean directory to ensure no DLL dependencies"