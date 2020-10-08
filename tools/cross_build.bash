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
  target=$1

  if [[ "$target" == "" ]]; then 
    echo ">>> usage: cross_build.bash [makefile target]"
    exit 1
  fi

  if [[ "$BUILDPLATFORM" != "linux/amd64" ]]; then
    echo ">>> ERROR: BUILDPLATFORM must be linux/amd64, got $BUILDPLATFORM"
    exit 1
  fi

  case "$TARGETPLATFORM" in
    linux/amd64)  native_build $target       ;;
    linux/arm64)  seego_build  $target arm64 ;;
    linux/arm/v7) seego_build  $target arm 7 ;;
    *)
      echo ">>> ERROR: unsupported TARGETPLATFORM value $TARGETPLATFORM"
      exit 1
  esac
}

native_build() {
  CROSS_BUILD=false make $1
}

seego_build() {
  echo ">>> building $1 for $2"
  CROSS_BUILD=true CGO_ENABLED=1 GOOS=linux GOARCH=$2 GOARM=$3 make $1
}

main $1
