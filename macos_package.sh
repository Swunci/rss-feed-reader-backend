#!/bin/bash
set -e

APP_NAME="RSS Reader"
BUNDLE_ID="com.swunci.rssreader"
VERSION="${1:-1.0}"
BINARY="${2:-rss-reader-macos}"
OUT_DIR="${3:-.}"
PNG="cmd/rss-feed-backend/desktop/tempmacicon.png"

APP_DIR="$OUT_DIR/$APP_NAME.app"
CONTENTS="$APP_DIR/Contents"
MACOS="$CONTENTS/MacOS"
RESOURCES="$CONTENTS/Resources"

mkdir -p "$MACOS" "$RESOURCES"

cp "$BINARY" "$MACOS/rss-reader"

# Convert PNG to icns
mkdir -p /tmp/icon.iconset
# Standard Sizes
sips -s format png -Z 512  --padToHeightWidth 512 512   "$PNG" --out /tmp/icon.iconset/icon_512x512.png
sips -s format png -Z 256  --padToHeightWidth 256 256   "$PNG" --out /tmp/icon.iconset/icon_256x256.png
sips -s format png -Z 128  --padToHeightWidth 128 128   "$PNG" --out /tmp/icon.iconset/icon_128x128.png
sips -s format png -Z 32   --padToHeightWidth 32 32     "$PNG" --out /tmp/icon.iconset/icon_32x32.png
sips -s format png -Z 16   --padToHeightWidth 16 16     "$PNG" --out /tmp/icon.iconset/icon_16x16.png

# Retina (@2x) Sizes
sips -s format png -Z 1024 --padToHeightWidth 1024 1024 "$PNG" --out /tmp/icon.iconset/icon_512x512@2x.png
sips -s format png -Z 512  --padToHeightWidth 512 512   "$PNG" --out /tmp/icon.iconset/icon_256x256@2x.png
sips -s format png -Z 256  --padToHeightWidth 256 256   "$PNG" --out /tmp/icon.iconset/icon_128x128@2x.png
sips -s format png -Z 64   --padToHeightWidth 64 64     "$PNG" --out /tmp/icon.iconset/icon_32x32@2x.png
sips -s format png -Z 32   --padToHeightWidth 32 32     "$PNG" --out /tmp/icon.iconset/icon_16x16@2x.png
iconutil -c icns /tmp/icon.iconset -o "$RESOURCES/icon.icns"
rm -rf /tmp/icon.iconset

cat > "$CONTENTS/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundleDisplayName</key>
    <string>$APP_NAME</string>
    <key>CFBundleIdentifier</key>
    <string>$BUNDLE_ID</string>
    <key>CFBundleVersion</key>
    <string>$VERSION</string>
    <key>CFBundleShortVersionString</key>
    <string>$VERSION</string>
    <key>CFBundleExecutable</key>
    <string>rss-reader</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleIconFile</key>
    <string>icon</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key>
    <true/>
</dict>
</plist>
PLIST

echo "Done! $APP_DIR is ready."
