#!/usr/bin/env bash
#
# cross_build.bash will set appropriate GOOS/GOARCH values given a TARGETPLATFORM
# variable. TARGETPLATFORM will accept values populated by docker buildx.
#
# $BUILDPLATFORM must be linux/amd64.
#
# Only the following values for $TARGETPLATFORM are supported:
#   linux/amd64
#   linux/arm64
#   linux/arm/v7

main() {
  if [[ "$BUILDPLATFORM" != "linux/amd64" ]]; then
    echo ">>> ERROR: BUILDPLATFORM must be linux/amd64, got $BUILDPLATFORM"
    exit 1
  fi

  case "$TARGETPLATFORM" in
    linux/amd64)  native_build      ;;
    linux/arm64)  seego_build arm64 ;;
    linux/arm/v7) seego_build arm 7 ;;
    *)
      echo ">>> ERROR: unsupported TARGETPLATFORM value $TARGETPLATFORM"
      exit 1
  esac
}

native_build() {
  make CROSS_BUILD=false agent
}

seego_build() {
  echo ">>> building for $1"
  CROSS_BUILD=true CGO_ENABLED=1 GOOS=linux GOARCH=$1 GOARM=$2 make agent
}

main

