---
aliases:
- /docs/grafana-cloud/agent/flow/tutorials/flow-by-example/faq/
- /docs/grafana-cloud/monitor-infrastructure/agent/flow/tutorials/flow-by-example/faq/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/flow/tutorials/flow-by-example/faq/
- /docs/grafana-cloud/send-data/agent/flow/tutorials/flow-by-example/faq/
canonical: https://grafana.com/docs/agent/latest/flow/tutorials/flow-by-example/faq/
description: Getting started with Flow-by-Example Tutorials
title: Get started
weight: 300
---

## Who is this for?

This set of tutorials contains a collection of examples that build on each other to demonstrate how to configure and use [{{< param "PRODUCT_NAME" >}}][flow]. It assumes you have a basic understanding of what {{< param "PRODUCT_ROOT_NAME" >}} is and telemetry collection in general. It also assumes a base level of familiarity with Prometheus and PromQL, Loki and LogQL, and basic Grafana navigation. It assumes no knowledge of {{< param "PRODUCT_NAME" >}} or River concepts.

[flow]: https://grafana.com/docs/agent/latest/flow

## What is Flow?

Flow is a new way to configure {{< param "PRODUCT_NAME" >}}. It is a declarative configuration language that allows you to define a pipeline of telemetry collection, processing, and output. It is built on top of the [River](https://github.com/grafana/river) configuration language, which is designed to be fast, simple, and debuggable.

## What do I need to get started?

You will need a Linux or Unix environment with Docker installed. The examples are designed to be run on a single host so that you can run them on your laptop or in a VM. You are encouraged to follow along with the examples using a `config.river` file and experiment with the examples yourself.

To run the examples, you should have a Grafana Agent binary available. You can follow the instructions on how to [Install Grafana Agent as a Standalone Binary](https://grafana.com/docs/agent/latest/flow/setup/install/binary/#install-grafana-agent-in-flow-mode-as-a-standalone-binary) to get a binary.

## How should I follow along?

You can <a href="../docker-compose.yaml" download="docker-compose.yaml">click here to download the docker-compose</a> file to run and play with the examples.

After running `docker-compose up`, open [http://localhost:3000](http://localhost:3000) in your browser to view the Grafana UI.

The tutorials are designed to be followed in order and generally build on each other. Each example explains what it does and how it works. They are designed to be run locally, so you can follow along and experiment with them yourself.

The Recommended Reading sections in each tutorial provide a list of documentation topics. To help you understand the concepts used in the example, read the recommended topics in the order given.
