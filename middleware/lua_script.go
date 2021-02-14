package middleware

// rateLimitScript is script for ratelimiter
const rateLimitScript = `
-- KEYS[1]: IP address
local IP = KEYS[1]
local timeNow = tonumber(ARGS[1])
local ipLimit = tonumber(ARGS[2])
local duration = tonumber(ARGS[3])

if redis.call('exists', IP) > 0 then
    local num = redis.call('incrby', IP, -1)
    if num < 0 then
        return -1
    else
        return num
    end
else
    redis.call('setex', IP, duration, ipLimit - 1)
end
`
