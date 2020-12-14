#!/usr/bin/env bash
# grafanacloud-install.sh installs the Grafana Cloud Agent on supported x86_64
# Linux systems for Grafana Cloud users. Those who aren't users of Grafana Cloud 
# or need to install the Agent on a different architecture or platform should 
# try another installation method.
#
# grafanacloud-install.sh has a hard dependency on being run on a supported
# Linux system. Currently only systems that can install deb or rpm packages 
# are supported. The target system will try to be detected, but if it cannot, 
# PACKAGE_SYSTEM can be passed as an environment variable with either rpm or 
# deb.
set -euo pipefail

log() {
  echo "$@" >&2 
}

fatal() {
  log "$@"
  exit 1
}

#
# REQUIRED environment variables. 
#
GCLOUD_STACK_ID=${GCLOUD_STACK_ID:=} # Stack ID where integrations are installed
GCLOUD_API_KEY=${GCLOUD_API_KEY:=}   # API key to communicate to the integrations API

[ -z "$GCLOUD_STACK_ID" ] && fatal "Required environment variable $$GCLOUD_STACK_ID not set."
[ -z "$GCLOUD_API_KEY" ]  && fatal "Required environment variable $$GCLOUD_API_KEY not set."

#
# OPTIONAL environment variables. 
#

# Package system to install the Agent with. If not empty, MUST be either rpm or
# deb. If empty, the script will try to detect the host OS and the appropriate
# package system to use.
PACKAGE_SYSTEM=${PACKAGE_SYSTEM:=}

# 
# Global constants.
#
RELEASE_VERSION="0.9.0"

RELEASE_URL="https://github.com/grafana/agent/releases/download/v${RELEASE_VERSION}"
DEB_URL="${RELEASE_URL}/grafana-agent-${RELEASE_VERSION}-1.x86_64.deb"
RPM_URL="${RELEASE_URL}/grafana-agent-${RELEASE_VERSION}-1.x86_64.rpm"

# 
main() {
  if [ -z "$PACKAGE_SYSTEM" ]; then 
    PACKAGE_SYSTEM=$(detect_package_system)
  fi
  log "--- Using package system $PACKAGE_SYSTEM. Downloading and installing package"

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

  log '--- Retrieving config and placing in /etc/grafana-agent.yaml'
  retrieve_config | sudo tee /etc/grafana-agent.yaml

  log '--- Enabling and starting grafana-agent.service'
  sudo systemctl enable grafana-agent.service
  sudo systemctl start grafana-agent.service 
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

# install_deb downloads and installs the deb package of the Grafana Cloud Agent.
install_deb() {
  curl -sL "${DEB_URL}" -o /tmp/grfana-agent.deb
  sudo dpkg -i /tmp/grafana-agent.deb
  rm /tmp/grafana-agent.deb
}

# install_rpm downloads and installs the deb package of the Grafana Cloud Agent.
install_rpm() {
  sudo rpm --reinstall "${RPM_URL}"
}

# retrieve_config downloads the config file for the Agent and prints out its
# contents to stdout.
retrieve_config() {
  grafana-agentctl cloud-config -u "${GCLOUD_STACK_ID}" -p "${GCLOUD_API_KEY}" || fatal 'Failed to retrieve config'
}

main
