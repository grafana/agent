---
title: Install Grafana Agent on macOS
weight: 130
---
## Install Grafana Agent on macOS

Install Grafana Agent and get it up and running on macOS. 

> **Note:** If you intend to ship your data to Grafana Cloud, you can set up Grafana Agent using a Grafana Cloud integration. See [how to install an integration](/docs/grafana-cloud/data-configuration/integrations/install-and-manage-integrations/) and details about the [macOS integration](/docs/grafana-cloud/data-configuration/integrations/integration-reference/integration-macos-node/). 

### Overview
Use Homebrew to install the most recent released version of Grafana using the Homebrew package. You can also install Grafana Agent on macOS using the macOS binary.

### Steps

1. Open a terminal and enter:
   
   ```
   brew update
   brew install grafana-agent
   ```
   
    The brew page downloads and enters the files into:
    - /usr/local/Cellar/grafana-agent/[version] (Homebrew v2)
    - /opt/homebrew/Cellar/grafana-agent/[version] (Homebrew v3)
    - Grafana Agent logs should be located in `/opt/homebrew/var/log/` though this path may differ depending on the version of Homebrew.

1. Run the following commands:

    ```
    mkdir -p $(brew --prefix)/etc/grafana-agent/
    touch $(brew --prefix)/etc/grafana-agent/config.yml
    ```

1. Modify `config.yml` with your configuration requirements. 
    
    ```
    vi $(brew --prefix)/etc/grafana-agent/config.yml
    ```

    See [Configure Grafana Agent for details]({{< relref "../configuration/" >}}). 
    
  
1. Start Grafana Agent using the command:

    ` brew services start grafana-agent`

    For logs, see:
    - stdout: `$(brew --prefix)/var/log/grafana-agent.log`
    - stderr: `$(brew --prefix)/var/log/grafana-agent.err.log`

1. Enter the following command to upgrade Grafana Agent:

    `brew upgrade grafana-agent`.



    


   
