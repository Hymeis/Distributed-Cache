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
## wrk2 BenchMark Testing
Sustained a constant 15 000 QPS workload with wrk2, observing:
- Throughput: 14 982 req/sec (≈15 000 target)
- Mean latency: 0.868 ms
- P50 / P75 / P90: 0.86 ms / 1.15 ms / 1.44 ms
- P99: 1.94 ms (well under 10 ms SLO)
- P99.9 / P99.99: 2.40 ms / 2.86 ms
- Max observed: 4.00 ms

---
## Next Steps
- Application side: Maybe make a Leetcode Top K ranking system? 
# How to run the project
Try
```
bash run.sh
```
and check the output shown