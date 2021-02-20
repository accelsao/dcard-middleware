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
Rate Limiting
1000 times in an hour for each IP address
response `X-RateLimit-Remaining` (remains time to visit) and `X-RateLimit-Reset` (time to reset)

- [X] count client visiting counts
- [X] add timeout for each IP address
- [X] add get/remove test (mem)
- [X] add get/remove test (Redis)
- [X] add Redis Ring (test only)
- [ ] use `PXAT` if `alicebob/miniredis` is available for redis v6.2
- [ ] add `atomic` for mem

# QA
## Why use Redis, instead of golang map?
`INCR` and `EXPIRE` are useful
> ref: https://redislabs.com/redis-best-practices/basic-rate-limiting/