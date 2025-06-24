# Distributed-Cache
A lightweight, Go-based distributed in-memory cache with  **consistent-hashing sharding**, **singleflight deduplication**, **size-bounded LRU**, **protobuf communication**, and **read-through replication**
## Architecture

```text
Client → Group.Add("🐺", "Hymeis")
        └─> Cache.Add("🐺", "Hymeis")
             ├─ insert into in-memory LRU
             └─ async fan-out to R-1 successors:
                  └─ for each replica in GetReplicas("🐺", R)[1:]:
                       HTTP POST /dcache/<group>/🐺  (SetRequest)

Client → Group.Get("🐺")
        ├─ LRU hit? ──▶ return "Hymeis"
        └─ cache miss:
            └─ singleflight.Do("🐺", fn):
                └─ pickPeer("🐺") via consistent-hash
                    ├─ peer? ──▶ peerLoad (HTTP+Protobuf) ──▶ return "Hymeis"
                    └─ local?  ──▶ localLoad:
                         ├─ GetterFunc → origin data
                         ├─ Replication() (see Add flow above)
                         └─ return "Hymeis"

```
---
## Features TBD
- Node faliure handling (via replication)
- [More]
# How to run the project
Try
```
bash run.sh
```
and check the output shown