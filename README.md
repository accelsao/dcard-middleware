# Usage
1. Run Redis Server (port: 6379)
2. Run main.go
3. Run client request
    * `curl -I 127.0.0.1:8080/hello`
    * `curl -H "X-Forwarded-For: 1.2.5.4:8080" -I 127.0.0.1:8080/hello`

# dcard-middleware
- [X] client visit times count
- [X] server timeout for each IP
- [X] add get/remove test (redis)
- [X] add get/remove test (mem)
- [ ] redis call, use set with exat or pxat after redis 6.2 is available
- [ ] add policy, iplimit/duration

# QA
## Why use Redis, instead of golang map?
`INCR` and `EXPIRE` are useful
> ref: https://redislabs.com/redis-best-practices/basic-rate-limiting/
## HSET or SET
TODO