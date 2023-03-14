# Architecture to install.
ARCH=${ARCH:=arm64}

#
# Global constants.
#
RELEASE_VERSION="v0.32.1"

# The DEB and RPM urls don't include the `v` version prefix in the file names,
# so we trim it out using ${RELEASE_VERSION#v} below.
RELEASE_URL="https://github.com/grafana/agent/releases/download/${RELEASE_VERSION}"

DEB_URL="${RELEASE_URL}/grafana-agent-${RELEASE_VERSION#v}-1.${ARCH}.deb"
RPM_URL="${RELEASE_URL}/grafana-agent-${RELEASE_VERSION#v}-1.${ARCH}.rpm"
DARWIN_URL="${RELEASE_URL}/grafana-agent-darwin-$ARCH.zip"

main() {
  if [ -z "$PACKAGE_SYSTEM" ]; then
    PACKAGE_SYSTEM=$(detect_package_system)
  fi
   echo "--- Using package system $PACKAGE_SYSTEM. Downloading and installing package for ${ARCH}"

  case "$PACKAGE_SYSTEM" in
    darwin)
      install_darwin
      ;;
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

   echo '--- Running autodiscovery for Grafana Agent and placing config in ~/agent-config.river ...'
  run_autodiscovery | tee ~/agent-config.river

   echo '--- You can now inspect the generated config and start the Agent using the following command.'
   echo '--- AGENT_MODE=flow ./grafana-agent run ~/agent-config.river'
   echo '---'
   echo '--- Lets navigate to Grafana Cloud to explore our data!'
}


# detect_package_system tries to detect the host distribution to determine if
# deb or rpm should be used for installing the Agent. Prints out either "deb"
# or "rpm". Calls fatal if the host OS is not supported.
detect_package_system() {
  command -v dpkg >/dev/null 2>&1 && { echo "deb"; return; }
  command -v rpm  >/dev/null 2>&1 && { echo "rpm"; return; }

  case "$(uname)" in
    Darwin)
      echo "darwin"; return;
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

# install_darwin downloads and installs the darwin binary of the Grafana Agent.
install_darwin() {
    curl -fL# "${DARWIN_URL}" -o /tmp/grafana-agent.zip || fatal 'Failed to download package'
    unzip /tmp/grafana-agent.zip -d /tmp
    mv /tmp/grafana-agent-darwin-arm64 grafana-agent
}


# run_autodiscovery generates the config file based on what is detected on the
# host system.
run_autodiscovery() {
  echo 'Running autodiscovery!'
  ./grafana-agent autodiscover
}

main

