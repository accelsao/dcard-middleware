# Usage
1. Run Redis Server (port: 6379)
2. Run main.go
3. Run client request
    * `curl -I 127.0.0.1:8080/hello`
    * `curl -H "X-Forwarded-For: 1.2.3.4:8080" -I 127.0.0.1:8080/hello`

# dcard-middleware
- [X] client visit times count
- [X] server timeout for each IP
- [ ] start server by golang `select` (better?)
- [X] add get/remove test (redis)
- [X] add get/remove test (mem)
- [ ] add concurrent test
- [ ] add main.go
- [ ] redis call, use set with exat or pxat after redis 6.2 is available
# QA
## Why use Redis, instead of golang map?
`INCR` and `EXPIRE` are useful
> ref: https://redislabs.com/redis-best-practices/basic-rate-limiting/