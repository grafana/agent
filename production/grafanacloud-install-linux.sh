#!/usr/bin/env sh
# shellcheck shell=dash
# This script should run in all POSIX environments and Dash is POSIX compliant.

# grafanacloud-install-linux.sh installs the Grafana Agent on supported
# Linux systems for Grafana Cloud users. Those who aren't users of Grafana Cloud
# or need to install the Agent on a different architecture or platform should
# try another installation method.
#
# grafanacloud-install.sh has a hard dependency on being run on a supported
# Linux system. Currently only systems that can install deb or rpm packages
# are supported. The target system will try to be detected, but if it cannot,
# PACKAGE_SYSTEM can be passed as an environment variable with either rpm or
# deb.
set -eu
trap "exit 1" TERM
MY_PID=$$

log() {
  echo "$@" >&2
}

fatal() {
  log "$@"
  kill -s TERM "$MY_PID"
}

#
# REQUIRED environment variables.
#

# None

#
# OPTIONAL environment variables.
#

# Architecture to install.
ARCH=${ARCH:=amd64}

# Package system to install the Agent with. If not empty, MUST be either rpm or
# deb. If empty, the script will try to detect the host OS and the appropriate
# package system to use.
PACKAGE_SYSTEM=${PACKAGE_SYSTEM:=}

#
# Global constants.
#
RELEASE_VERSION="v0.28.0"

# The DEB and RPM urls don't include the `v` version prefix in the file names,
# so we trim it out using ${RELEASE_VERSION#v} below.
RELEASE_URL="https://github.com/grafana/agent/releases/download/${RELEASE_VERSION}"
DEB_URL="${RELEASE_URL}/grafana-agent-${RELEASE_VERSION#v}-1.${ARCH}.deb"
RPM_URL="${RELEASE_URL}/grafana-agent-${RELEASE_VERSION#v}-1.${ARCH}.rpm"

main() {
  if [ -z "$PACKAGE_SYSTEM" ]; then
    PACKAGE_SYSTEM=$(detect_package_system)
  fi
  log "--- Using package system $PACKAGE_SYSTEM. Downloading and installing package for ${ARCH}"

  case "$PACKAGE_SYSTEM" in
    deb)
      install_deb
      ;;
    rpm)
      install_rpm
      ;;
    *)
      fatal "Unsupported PACKAGE_SYSTEM value $PACKAGE_SYSTEM. Must be either rpm or deb".
      ;;
  esac

  log '--- Enabling grafana-agent.service'
  sudo systemctl enable grafana-agent.service

  # Add some empty newlines to give some visual whitespace before printing the
  # success message.
  log ''
  log ''
  log 'Grafana Agent is now installed! To check the status of your Agent, run:'
  log '   sudo systemctl status grafana-agent.service'
}

# detect_package_system tries to detect the host distribution to determine if
# deb or rpm should be used for installing the Agent. Prints out either "deb"
# or "rpm". Calls fatal if the host OS is not supported.
detect_package_system() {
  command -v dpkg >/dev/null 2>&1 && { echo "deb"; return; }
  command -v rpm  >/dev/null 2>&1 && { echo "rpm"; return; }

  case "$(uname)" in
    Darwin)
      fatal 'macOS not supported'
      ;;
    *)
      fatal "Unknown unsupported OS: $(uname)"
      ;;
  esac
}

# install_deb downloads and installs the deb package of the Grafana Agent.
install_deb() {
  curl -fL# "${DEB_URL}" -o /tmp/grafana-agent.deb || fatal 'Failed to download package'
  sudo dpkg -i /tmp/grafana-agent.deb
  rm /tmp/grafana-agent.deb
}

# install_rpm downloads and installs the deb package of the Grafana Agent.
install_rpm() {
  sudo rpm --reinstall "${RPM_URL}"
}

main
