---
aliases:
- /docs/agent/shared/flow/reference/components/otelcol-grpc-balancer-name
headless: true
---

The supported values for `balancer_name` are listed in the gRPC documentation on [Load balancing][]:
* `pick_first`: Tries to connect to the first address, uses it for all RPCs if it connects, 
  or tries the next address if it fails (and keep doing that until one connection is successful). 
  Because of this, all the RPCs will be sent to the same backend.
* `round_robin`: Connects to all the addresses it sees, and sends an RPC to each backend one at a time in order. 
  E.g. the first RPC will be sent to backend-1, the second RPC will be be sent to backend-2, 
  and the third RPC will be be sent to backend-1 again.

[Load balancing]: https://github.com/grpc/grpc-go/blob/master/examples/features/load_balancing/README.md#pick_first
