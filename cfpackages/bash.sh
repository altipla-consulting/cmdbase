#!/bin/sh

set -eu

VERSION=$(curl -s https://packages.altipla.consulting/{{.App}}/stable.txt)

if [ "${GITHUB_ACTIONS:-}" = "true" ]; then
  INSTALL_DIR="$RUNNER_TEMP/bin"
  echo "$INSTALL_DIR" >> "$GITHUB_PATH"
else
  INSTALL_DIR="$HOME/bin"
fi
if [ ! -d "$INSTALL_DIR" ]; then
  mkdir -p "$INSTALL_DIR"
fi

curl -q --fail --location --progress-bar --output /tmp/install.tar.gz "https://packages.altipla.consulting/bin/{{.App}}/{{.App}}-linux-amd64-$VERSION.tar.gz"
cd "$INSTALL_DIR"
tar xzf /tmp/install.tar.gz
rm /tmp/install.tar.gz

echo "{{.App}} $VERSION was installed successfully!"
if command -v {{.App}} >/dev/null; then
  echo "Run '{{.App}} help' to get started."
else
  "$INSTALL_DIR/{{.App}}" install
fi
