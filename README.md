# Usage
1. Run Redis Server (port: 6379)
2. Run main.go
    Configuration
    * useMem(Bool): set if you dont use Redis
    * ipLimit(Int): rate limit for certain IP
    * duration(time.Duration): time for rate limiting, reset after timeout
    
3. Run client request
    * `curl -I 127.0.0.1:8080/hello`
    * `curl -H "X-Forwarded-For: 1.2.5.4:8080" -I 127.0.0.1:8080/hello`

## Example
`go run main.go -duration 1s -ipLimit 2`

![](https://i.imgur.com/R2zeoWi.png)

# dcard-middleware
- [X] client visit times count
- [X] server timeout for each IP
- [X] add get/remove test (redis)
- [X] add get/remove test (mem)
- [ ] redis call, use set with exat or pxat after redis 6.2 is available

# QA
## Why use Redis, instead of golang map?
`INCR` and `EXPIRE` are useful
> ref: https://redislabs.com/redis-best-practices/basic-rate-limiting/