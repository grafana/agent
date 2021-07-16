#!/usr/bin/env bash
#
# update_formula.bash will patch the Grafana Agent formula in the checked out
# tap directory to install the specified version.
set -eo pipefail

FORMULA_PATH="grafana-cloud-agent.rb"
TAP_DIR=""
VERSION=""

for arg in "$@"; do
  case $arg in
    -v|--version)
      VERSION="$2"
      shift; shift # Remove name and value
      ;;
    -t|--tap-dir)
      TAP_DIR="$2"
      shift; shift # Remove name and value
      ;;
  esac
done

if [ -z "$TAP_DIR" ] || [ -z "$VERSION" ]; then
  echo "usage: $0 --tap-dir <path to tap> --version <new version>"
  exit 1
fi

NEW_URL="https://github.com/grafana/agent/archive/${VERSION}.tar.gz"
NEW_SHA=$(curl -fsSL "$NEW_URL" | sha256sum | awk '{print $1}')

echo "New URL: $NEW_URL"
echo "New SHA: $NEW_SHA"

sed -E 's#url "(.*)"#url "'$NEW_URL'"#g; s#sha256 "(.*)"#sha256 "'$NEW_SHA'"#g' \
  "$TAP_DIR/$FORMULA_PATH" > $FORMULA_PATH.tmp
mv $FORMULA_PATH.tmp "$TAP_DIR/$FORMULA_PATH"
