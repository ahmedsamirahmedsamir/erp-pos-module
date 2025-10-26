#!/bin/bash
set -e

echo "Building POS Module Plugin..."

case "$(uname -s)" in
    Linux*)     EXT=".so";;
    Darwin*)    EXT=".dylib";;
    MINGW*|MSYS*|CYGWIN*) EXT=".dll";;
    *)          EXT=".so";;
esac

cd handlers
go build -buildmode=plugin -o ../pos${EXT} plugin.go pos_handler_simple.go types.go
echo "✓ Plugin built successfully: pos${EXT}"

