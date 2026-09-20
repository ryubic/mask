#!/usr/bin/env bash

set -e

VERSION="1.0.0"

DIST_DIR="dist"
BUILD_DIR="build_tmp"

# 1. Check if required build tools exist

if ! command -v go >/dev/null 2>&1; then
    echo "Error: Go is not installed or not in PATH."
    exit 1
fi

if ! command -v zip >/dev/null 2>&1; then
    echo "Error: 'zip' utility is not installed."
    exit 1
fi

if ! command -v sha256sum >/dev/null 2>&1; then
    echo "Error: 'sha256sum' utility is not installed."
    exit 1
fi

# 2. Ensure config.json exists

if [ ! -f "config.json" ]; then
    printf '\033[1;31m[mask error] Could not load config.json: file not found.\033[0m\n' >&2
    exit 1
fi

# 3. Reset distribution directory

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# 4. Target Matrix (GOOS/GOARCH pairs)

PLATFORMS=(
    "windows amd64"
    "windows arm64"
    "darwin amd64"
    "darwin arm64"
    "linux amd64"
    "linux arm64"
)

echo "================================================="
echo " Building mask & cmask v${VERSION} for all targets"
echo "================================================="

for PLATFORM in "${PLATFORMS[@]}"; do
    read -r GOOS GOARCH <<< "$PLATFORM"

    MASK_BIN="mask"
    CMASK_BIN="cmask"

    if [ "$GOOS" = "windows" ]; then
        MASK_BIN="mask.exe"
        CMASK_BIN="cmask.exe"
    fi

    PACKAGE_NAME="mask-v${VERSION}-${GOOS}-${GOARCH}"
    ZIP_NAME="${PACKAGE_NAME}.zip"

    echo "--> Compiling mask and cmask for ${GOOS}/${GOARCH}..."

    # Reset temporary staging directory

    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR/$PACKAGE_NAME"

    # Compile stripped binary for mask

    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build \
        -ldflags="-s -w" \
        -o "$BUILD_DIR/$PACKAGE_NAME/$MASK_BIN" \
        mask.go

    # Compile stripped binary for cmask

    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build \
        -ldflags="-s -w" \
        -o "$BUILD_DIR/$PACKAGE_NAME/$CMASK_BIN" \
        cmask.go

    # Copy config file

    cp config.json "$BUILD_DIR/$PACKAGE_NAME/"

    # Create ZIP with package directory as the root

    (
        cd "$BUILD_DIR"
        zip -q -r "../$DIST_DIR/$ZIP_NAME" "$PACKAGE_NAME"
    )

    # Cleanup

    rm -rf "$BUILD_DIR"

    echo "    Created $DIST_DIR/$ZIP_NAME"
done

# 5. Generate SHA-256 checksums

echo ""
echo "--> Generating SHA-256 checksums..."

(
    cd "$DIST_DIR"
    sha256sum *.zip > SHA256SUMS.txt
)

echo "    Created $DIST_DIR/SHA256SUMS.txt"

# 6. Final output

echo ""
echo "================================================="
echo " All packages built successfully in '${DIST_DIR}/'"
echo "================================================="

ls -lh "$DIST_DIR"
