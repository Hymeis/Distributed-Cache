# Distributed-Cache
Learning Golang through a distributed project
## Features Done
- LRU Cache
- Local Cache CRUD (w/ pessimistic lock)
## Features TBD
- Distributed Cache CRUD
- Consistent Hashing
- [More]
# How to run the project
```
go build
./distributed-cache
```
Try (for example)
```
curl http://localhost:8080/cache/scores/Lang
```
In another terminal and see what happens 