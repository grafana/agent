---
title: Other systems
weight: 600
aliases:
 - /docs/sources/flow/install/binary/
---

# Run Grafana Agent Flow on other systems

Grafana Agent Flow is distributed as plain binaries for various systems:

* Linux: AMD64, ARM64, ARMv6, ARMv7
* Windows: AMD64
* macOS: AMD64 (Intel), ARM64 (Apple Silicon)
* FreeBSD: AMD64

## Steps

1. Download Grafana Agent:

   1. Navigate to the current Grafana Agent [release][] page.

   2. Scroll down to the **Assets** section.

   3. Download the version that matches your operating system and machine's
      architecture.

   4. Extract the package contents into a directory.

   5. If running on Linux, macOS, or FreeBSD, run the following command in a
      terminal:

      ```bash
      chmod +x BINARY_PATH
      ```

      Replace `BINARY_PATH` with the path to the extracted binary from step 1.4.

2. Run Grafana Agent Flow:

   1. If running on Linux, macOS, or FreeBSD, run the following command in a
      terminal:

      ```bash
      AGENT_MODE=flow BINARY_PATH run CONFIG_FILE
      ```

      * Replace `BINARY_PATH` with the path to the extracted binary from step
        1.4.
      * Replace `CONFIG_FILE` with the path to a Grafana Agent Flow
        configuration file to run.

   2. If running on Windows, run the following in a command prompt:

      ```cmd
      set AGENT_MODE=flow
      BINARY_PATH run CONFIG_FILE
      ```

      * Replace `BINARY_PATH` with the path to the extracted binary from step
        1.4.
      * Replace `CONFIG_FILE` with the path to a Grafana Agent Flow
        configuration file to run.

[release]: https://github.com/grafana/agent/releases/latest

