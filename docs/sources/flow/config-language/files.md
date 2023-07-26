---
aliases:
- ../configuration-language/files/
canonical: https://grafana.com/docs/agent/latest/flow/config-language/files/
title: Files
weight: 100
---

# Files
River files are plaintext files with the `.river` file extension. Each River
file may be referred to as a "configuration file," or a "River configuration."

River files are required to be UTF-8 encoded, and are permitted to contain
Unicode characters. River files can use both Unix-style line endings (LF) and
Windows-style line endings (CRLF), but formatters may replace all line endings
with Unix-style ones.
