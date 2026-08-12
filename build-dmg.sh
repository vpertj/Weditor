#!/usr/bin/env bash
# WEditor 一键打包 DMG
# 用法: ./build-dmg.sh
set -euo pipefail

cd "$(dirname "$0")"

VERSION=$(grep -o '"productVersion": *"[^"]*"' wails.json | sed 's/.*"\([^"]*\)"$/\1/')
DMG_NAME="WEditor-${VERSION}.dmg"

echo "=== [1/3] wails build ==="
wails build

echo "=== [2/3] 生成 DMG ==="
STAGE=$(mktemp -d /tmp/WEditor-dmg-stage.XXXXXX)
trap 'rm -rf "$STAGE"' EXIT
cp -R build/bin/WEditor.app "$STAGE/"
ln -s /Applications "$STAGE/Applications"

cd build/bin
rm -f "$DMG_NAME"
hdiutil create -volname "WEditor" -srcfolder "$STAGE" -ov -format UDZO "$DMG_NAME"

echo "=== [3/3] 完成 ==="
echo "📦 $(pwd)/$DMG_NAME"
