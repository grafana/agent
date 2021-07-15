#!/usr/bin/env bash

FOUND_CC=""
FOUND_CXX=""
FOUND_LD_PATH=""
CC_EXTRA=""
CXX_EXTRA=""

CGO_ENABLED=${CGO_ENABLED:-$(go env CGO_ENABLED)}
GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)
GOARM=$(go env GOARM)

main() {
  echo ">>> discovering toolchain for GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM}"

  case "$GOOS" in
    linux)   configure_linux   ;;
    darwin)  configure_darwin  ;;
    freebsd) configure_bsd     ;;
    windows) configure_windows ;;
    *)
      echo ">>> ERROR: unsupported GOOS value $GOOS"
      exit 1
  esac

  echo ">>> discovered CC ${FOUND_CC}"
  echo ">>> discovered CXX ${FOUND_CXX}"
  echo ">>> extra C flags: ${CC_EXTRA}"

  echo ">>> CGO_ENABLED=${CGO_ENABLED} CC=$FOUND_CC CXX=$FOUND_CXX go $@"

  # If we run go directly, any files created on the bind mount
  # will have awkward ownership.  So we switch to a user with the
  # same user and group IDs as source directory.  We have to set a
  # few things up so that sudo works without complaining later on.
  uid=$(stat --format="%u" $(pwd))
  gid=$(stat --format="%g" $(pwd))
  echo "seego:x:$uid:$gid::$(pwd):/bin/bash" >>/etc/passwd
  echo "seego:*:::::::" >>/etc/shadow
  echo "seego  ALL=(ALL) NOPASSWD: ALL" >>/etc/sudoers

  # I'm skeptical that this is the best way to do it because it's so
  # annoying, but it works at least.
  #
  # Export environment variables we want to retain when running go
  # as the new user. Even though we preserve the path, we have to use
  # the full path to go since sudo won't read the PATH for it.

  export CC="$FOUND_CC $CC_EXTRA"
  export CXX="$FOUND_CXX $CC_EXTRA"

  export CGO_CFLAGS=$CC_EXTRA
  export CGO_CXX_CFLAGS=$CC_EXTRA
  export CGO_LDFLAGS=$CC_EXTRA

  export PATH=$PATH
  if [ ! -z "$FOUND_LD_PATH" ]; then
    export LD_LIBRARY_PATH="$FOUND_LD_PATH:$LD_LIBRARY_PATH"
  fi

  export GOCACHE=$(pwd)/.cache

  exec sudo -E \
    --preserve-env=PATH \
    --preserve-env=LD_LIBRARY_PATH \
    -u seego -- $(which go) "$@"
}

configure_linux() {
  toolchain_prefix=""

  case "$GOARCH" in
    # Do nothing for native archs
    amd64 | 386) ;;

    arm)      toolchain_prefix="arm-linux-gnueabi-" ;;
    arm64)    toolchain_prefix="aarch64-linux-gnu-" ;;

    ppc64)    toolchain_prefix="powerpc-linux-gnu-"     ;;
    ppc64le)  toolchain_prefix="powerpc64le-linux-gnu-" ;;

    mips)     toolchain_prefix="mips-linux-gnu-"          ;;
    mipsle)   toolchain_prefix="mipsel-linux-gnu-"        ;;
    mips64)   toolchain_prefix="mips64-linux-gnuabi64-"   ;;
    mips64le) toolchain_prefix="mips64el-linux-gnuabi64-" ;;

    s390x)    toolchain_prefix="s390x-linux-gnu-" ;;

    *)
      echo ">>> ERROR: unsupported linux GOARCH value $GOARCH"
      echo ">>> supported values: amd64, 386, arm, arm64, ppc64, ppc64le, mips, mipsle, mips64, mips64le, s390x"
      exit 1
  esac

  if [ "$GOARCH" == "arm" ] && [ "$GOARM" == "7" ]; then
    toolchain_prefix="arm-linux-gnueabihf-"
  fi

  FOUND_CC="${toolchain_prefix}gcc"
  FOUND_CXX="${toolchain_prefix}g++"
}

configure_darwin() {
  case "$GOARCH" in
    amd64)
      FOUND_CC="x86_64-apple-darwin20.2-clang"
      FOUND_CXX="x86_64-apple-darwin20.2-clang++"
      FOUND_LD_PATH="$OSXCROSS_PATH/lib"
      ;;
    arm64)
      FOUND_CC="arm64-apple-darwin20.2-clang"
      FOUND_CXX="arm64-apple-darwin20.2-clang++"
      FOUND_LD_PATH="$OSXCROSS_PATH/lib"
      ;;
    *)
      echo ">>> ERROR: unsupported darwin GOARCH value $GOARCH"
      echo ">>> supported values: amd64, arm64"
      exit 1
  esac
}

configure_bsd() {
  case "$GOARCH" in
    amd64)
      FOUND_CC="clang"
      FOUND_CXX="clang++"
      CC_EXTRA="-target x86_64-pc-freebsd11 --sysroot=/usr/freebsd/x86_64-pc-freebsd11"
      ;;
    386)
      FOUND_CC="clang"
      FOUND_CXX="clang++"
      CC_EXTRA="-target i386-pc-freebsd11 --sysroot=/usr/freebsd/i386-pc-freebsd11 -v"
      ;;

    *)
      echo ">>> ERROR: unsupported bsd GOARCH value $GOARCH"
      echo ">>> supported values: amd64, 386"
      exit 1
  esac
}

configure_windows() {
  case "$GOARCH" in
    386)
      FOUND_CC="i686-w64-mingw32-gcc"
      FOUND_CXX="i686-w64-mingw32-g++"
      ;;
    amd64)
      FOUND_CC="x86_64-w64-mingw32-gcc"
      FOUND_CXX="x86_64-w64-mingw32-g++"
      ;;

    *)
      echo ">>> ERROR: unsupported windows GOARCH value $GOARCH"
      echo ">>> supported values: 386, amd64"
      exit 1
  esac
}

main "$@"