# dcard-middleware
- [X] client visit times count
- [X] server timeout for each IP
- [ ] start server by golang `select` (better?)
- [X] add get/remove test (redis)
- [X] add get/remove test (mem)
- [ ] add concurrent test
- [ ] add main.go

# QA
## Why use Redis, instead of golang map?
`INCR` and `EXPIRE` are useful
