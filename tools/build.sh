#!/bin/bash

# Default values
OUT_FILE="./tmp/main"
RUN_APP=false

# Parse arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -o|--output) OUT_FILE="$2"; shift ;;
        --run) RUN_APP=true ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

echo "Building CY_BORGER to ${OUT_FILE}..."

echo "generating static files..."
# Generate minified CSS
npm run build:css

echo "generating templ files..."
templ generate

echo "generating go binary..."
go build -o "$OUT_FILE" cmd/cy_borger/main.go

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build successful: $OUT_FILE"

# Run if requested
if [ "$RUN_APP" = true ]; then
    echo "Running $OUT_FILE..."
    exec "$OUT_FILE"
fi
