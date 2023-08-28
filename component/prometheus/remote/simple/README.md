# Simple Overview

Simple (name could be better), is meant to be a less complex WAL. The underlying concept is what if the WAL was a durable queue. How simple could we make it? 

## Goals

* Simplify code base
* Reduce memory footprint
* Replay by default
* Maintain rough performance parity with normal WAL
* Uncouple from scrape manager
* Strong TTL guarantees

## Overview

The core of each Simple component is a pair of [PebbleDB](https://github.com/cockroachdb/pebble) databases. These databases hold all the information needed for a single Simple component. 

The samples database consists of:
* Key - autoincrementing uint64
* Value - Item - Gob Encoded
    * Type - int8 value used to determine what sort of data is contained
    * TTL - Unix timestamp for the TTL expiration, 0 meaning no TTL
    * Data - Snappy and Gob encoded value

The bookmark database consists of:
* Key - string name that is unique pair of writer and remote write. (This is mainly for if we ever want to define multiple endpoints)
* Value - The last fully processed Key from the above.

The general flow of is below.

[Appender] -> [SimpleDB] -> [DBStore] -> [PebbleDB] -> [Writer] -> [QueueManager] -> [WriteClient] -> [Remote Write Endpoint]

### Parts

#### Simple

Simple is the underlying component that ties everything together. Simple primary goal is to return Appenders and to move data from the Appenders.Commit to the DBStore that then writes to the PebbleDB.

#### DBStore

DBStore wraps the sample and bookmark databases along with the KeyCache and handles high level operations, translating them into lower level operations in pebble/db. 

#### Pebble/DB

Pebble/DB allows low level access to the underlying pebble database. Keeps the current index key and provides primitives for Writing and Getting Values. All the data is Gob encoded and Snappy compressed. The lower level handles eviction and TTL. TTL is checked on retrieval and eviction. This means that even if the data exists on disk if it is expired it will not be returned thus allowing hard TTL promises. Passed into Pebble/DB are converters for converting to and from data structures. GetType and GetValue.

#### KeyCache

KeyCache is a memory cache of all the values added to the sample DB. This allows for quick checks without having to hit the disk. Though it is unbounded it only contains a few fields for each scrape. Key, TTL, Size.

#### Writer

Writer polls the DBStore for values and sends them to the QueueManager. The loop for handling data is likely the most in need of attention. Handling errors from the QueueManager causes decision making. For instance there is a check for out of order samples or to old that are considered non-recoverable. It also fails or succeeds the whole scrape in one go, instead the QueueManager should return the failed scrapes. 

#### QueueManager

QueueManager started life as a fork of prometheus/QueueManager. Its gone off the rails a bit. Much of the QueueManager code has been removed, the sharding logic has been simplified and hard coded to 4 shards. This shard limit is likely a todo to change. The other big change is shards are not long running but created on each request. This simplifies the code and makes it easier to return error codes specific to that request.

#### WriteClient

WriteClient has been pulled from prometheus with minimal changes.


## Performance

Benchmarking can be run with `PROM_USERNAME=<id> PROM_PASSWORD="<password"> ./bench.sh` in the benchmark directory. This will run a series of 6 tests each an hour long. A http server will be started that can return an error code on metrics push or consume them. An avalanche instance is started serving the specified number of metrics. Then 2 agents one using the new Remote and the other using the old Remote are spun up with Label rules adding `allow_wal` and `metric_count`.

The first test is 1,000 metrics and the endpoint consumes the wal.
The second test is 1,000 metrics and the endpoint returns an error.

The third and fourth are the same but use 10,000 samples.
The fifth and sixth are the same but use 100,000 samples.

#### Compression Ratio

For the avalanche examples the compression ratio is 16x. 

#### Memory Usage

Memory usage in simple stays relatively flat, it is primarily driven by the size of the largest scrape. 

#### CPU Usage

CPU usage of Simple is generally half of old.

#### Disk Usage

Disk storage is roughly 25% higher and disk IO is roughly 4x higher.
