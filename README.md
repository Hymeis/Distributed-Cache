# Distributed-Cache
A lightweight, Go-based distributed in-memory cache with  **Consistent-hashing sharding**, **Singleflight deduplication**, **Size-bounded LRU**, and **Protobuf Communcation**
## Architecture

```text
Client → Cache.Add("🐺", "Hymeis") (evict LRU if needed)
Client → Group.Get("🐺")
        ├─ LRU hit? ──▶ return
        └─ cache miss:
            └─ singleflight.Do("foo", fn): 
                └─ pickPeer("foo") via consistent hash
                    ├─ peer? ──▶ peerLoad (HTTP+Protobuf) ──▶ return
                    └─ local?  ──▶ localLoad ──▶ return
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