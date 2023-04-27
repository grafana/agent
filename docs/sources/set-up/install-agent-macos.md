---
title: Install Grafana Agent on macOS
weight: 130
---
## Install Grafana Agent on macOS

Install Grafana Agent and get it up and running on macOS. 

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
  
2. Start Grafana Agent using the command:

    ` brew services start grafana-agent`

    The Grafana Agent configuration file can be found at `$(brew --prefix)/etc/grafana-agent/config.yml`.

3. Enter the following command to upgrade Grafana Agent:

    `brew upgrade grafana-agent`.



    


   